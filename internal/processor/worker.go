package processor

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"segmentation-api/internal/models"
	"segmentation-api/internal/repository"
	"segmentation-api/internal/service"
)

type record struct {
	userID  uint64
	segType string
	name    string
	data    []byte
}

func Run(ctx context.Context, svc *service.SegmentationService, logger *log.Logger) error {
	filepath := os.Getenv("DATAFILEPATH")
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.FieldsPerRecord = -1

	// discard header
	if _, err := reader.Read(); err != nil {
		return err
	}

	workers := runtime.NumCPU()
	ch := make(chan record, workers*4)

	var (
		wg              sync.WaitGroup
		totalRead       uint64 // linhas lidas do CSV
		totalEnqueued   uint64 // registros válidos enviados ao channel
		totalProcessed  uint64 // registros inseridos
		totalFailed     uint64
		totalInvalid    uint64
		totalUpdated    uint64 // registros atualizados (duplicados)
		totalDuplicates uint64 // no-op duplicatas
		startTime       = time.Now()
		doneCh          = make(chan struct{})
	)

	// ─────────────────────────────────────────────
	// Progress reporter (fora do hot path)
	// ─────────────────────────────────────────────
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				read := atomic.LoadUint64(&totalRead)
				enq := atomic.LoadUint64(&totalEnqueued)
				ok := atomic.LoadUint64(&totalProcessed)
				upd := atomic.LoadUint64(&totalUpdated)
				dup := atomic.LoadUint64(&totalDuplicates)
				fail := atomic.LoadUint64(&totalFailed)
				invalid := atomic.LoadUint64(&totalInvalid)

				if read == 0 {
					continue
				}

				elapsed := time.Since(startTime).Seconds()
				rate := float64(ok+upd+dup) / elapsed

				logger.Printf(
					"progress read=%d enqueued=%d inserted=%d updated=%d duplicates=%d failed=%d invalid=%d rate=%.1f rec/s elapsed=%.fs",
					read, enq, ok, upd, dup, fail, invalid, rate, elapsed,
				)
			case <-doneCh:
				return
			case <-ctx.Done():
				logger.Println("processor_context_cancelled")
				return
			}
		}
	}()

	// ─────────────────────────────────────────────
	// Workers
	// ─────────────────────────────────────────────
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for r := range ch {
				select {
				case <-ctx.Done():
					return
				default:
				}

				seg := models.Segmentation{
					UserID:           r.userID,
					SegmentationType: r.segType,
					SegmentationName: r.name,
					Data:             r.data,
				}
				result, err := svc.Create(ctx, &seg)
				if err != nil {
					atomic.AddUint64(&totalFailed, 1)
					logger.Printf(
						"upsert_error worker=%d user_id=%d seg_type=%s seg_name=%s err=%v",
						workerID,
						r.userID,
						r.segType,
						r.name,
						err,
					)
					continue
				}

				switch result {
				case repository.UpsertInserted:
					atomic.AddUint64(&totalProcessed, 1)
					logger.Printf(
						"upsert_inserted worker=%d user_id=%d seg_type=%s seg_name=%s",
						workerID,
						r.userID,
						r.segType,
						r.name,
					)

				case repository.UpsertUpdated:
					atomic.AddUint64(&totalUpdated, 1)
					logger.Printf(
						"upsert_updated worker=%d user_id=%d seg_type=%s seg_name=%s",
						workerID,
						r.userID,
						r.segType,
						r.name,
					)

				case repository.UpsertNoOp:
					atomic.AddUint64(&totalDuplicates, 1)
					logger.Printf(
						"upsert_noop worker=%d user_id=%d seg_type=%s seg_name=%s",
						workerID,
						r.userID,
						r.segType,
						r.name,
					)
				}
			}
		}(i)
	}

	// ─────────────────────────────────────────────
	// Producer
	// ─────────────────────────────────────────────
	rowNum := 1 // header já descartado

	for {
		select {
		case <-ctx.Done():
			logger.Println("producer_context_cancelled")
			goto finish
		default:
		}

		row, err := reader.Read()
		rowNum++

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			logger.Printf("csv_read_error row=%d err=%v", rowNum, err)
			continue
		}

		atomic.AddUint64(&totalRead, 1)

		if len(row) < 4 {
			atomic.AddUint64(&totalInvalid, 1)
			logger.Printf("invalid_row_size row=%d size=%d", rowNum, len(row))
			continue
		}

		userID, err := strconv.ParseUint(strings.TrimSpace(row[0]), 10, 64)
		if err != nil {
			atomic.AddUint64(&totalInvalid, 1)
			logger.Printf("invalid_user_id row=%d value=%q", rowNum, row[0])
			continue
		}

		raw := strings.TrimSpace(row[3])
		if !json.Valid([]byte(raw)) {
			atomic.AddUint64(&totalInvalid, 1)
			logger.Printf("invalid_json row=%d", rowNum)
			continue
		}

		atomic.AddUint64(&totalEnqueued, 1)

		ch <- record{
			userID:  userID,
			segType: strings.TrimSpace(row[1]),
			name:    strings.TrimSpace(row[2]),
			data:    []byte(raw),
		}
	}

finish:
	close(ch)
	wg.Wait()
	close(doneCh)

	elapsed := time.Since(startTime)

	logger.Printf(
		"processor_finished read=%d enqueued=%d inserted=%d updated=%d duplicates=%d failed=%d invalid=%d elapsed=%s",
		totalRead,
		totalEnqueued,
		totalProcessed,
		totalUpdated,
		totalDuplicates,
		totalFailed,
		totalInvalid,
		elapsed.String(),
	)

	return nil
}
