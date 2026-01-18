package workers

import (
	"log"
	"sync"
	"time"

	"url-shortener/models"
	"url-shortener/repository"
)

// ClickAnalyticsWorker xử lý click events bất đồng bộ
// Sử dụng Goroutines và Channels để không làm chậm request chính
type ClickAnalyticsWorker struct {
	eventChannel  chan *models.ClickEvent
	urlRepo       *repository.URLRepositoryImpl
	analyticsRepo *repository.AnalyticsRepositoryImpl
	workerCount   int
	batchSize     int
	flushInterval time.Duration
	wg            sync.WaitGroup
	quit          chan struct{}
	isRunning     bool
	mu            sync.Mutex
}

// NewClickAnalyticsWorker tạo worker mới
func NewClickAnalyticsWorker(
	urlRepo *repository.URLRepositoryImpl,
	analyticsRepo *repository.AnalyticsRepositoryImpl,
	workerCount int,
	bufferSize int,
) *ClickAnalyticsWorker {
	return &ClickAnalyticsWorker{
		eventChannel:  make(chan *models.ClickEvent, bufferSize),
		urlRepo:       urlRepo,
		analyticsRepo: analyticsRepo,
		workerCount:   workerCount,
		batchSize:     100,             // Batch 100 events
		flushInterval: 5 * time.Second, // Flush mỗi 5 giây
		quit:          make(chan struct{}),
		isRunning:     false,
	}
}

// Start khởi động worker pool
func (w *ClickAnalyticsWorker) Start() {
	w.mu.Lock()
	if w.isRunning {
		w.mu.Unlock()
		return
	}
	w.isRunning = true
	w.mu.Unlock()

	log.Printf("🚀 Starting %d analytics workers...", w.workerCount)

	// Khởi động nhiều workers (Goroutines)
	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.worker(i)
	}

	log.Println("✅ Analytics workers started successfully")
}

// Stop dừng tất cả workers gracefully
func (w *ClickAnalyticsWorker) Stop() {
	w.mu.Lock()
	if !w.isRunning {
		w.mu.Unlock()
		return
	}
	w.mu.Unlock()

	log.Println("🛑 Stopping analytics workers...")

	// Đóng quit channel để signal stop
	close(w.quit)

	// Đợi tất cả workers hoàn thành
	w.wg.Wait()

	// Đóng event channel
	close(w.eventChannel)

	w.mu.Lock()
	w.isRunning = false
	w.mu.Unlock()

	log.Println("✅ Analytics workers stopped")
}

// Enqueue thêm event vào queue (non-blocking)
func (w *ClickAnalyticsWorker) Enqueue(event *models.ClickEvent) {
	// Non-blocking send với select
	select {
	case w.eventChannel <- event:
		// Event được enqueue thành công
	default:
		// Channel đầy, log warning nhưng không block
		log.Printf("⚠️ Analytics queue full, dropping event for: %s", event.ShortCode)
	}
}

// worker xử lý events từ channel
func (w *ClickAnalyticsWorker) worker(id int) {
	defer w.wg.Done()

	log.Printf("Worker %d started", id)

	batch := make([]*models.ClickEvent, 0, w.batchSize)
	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.quit:
			// Flush remaining batch trước khi thoát
			if len(batch) > 0 {
				w.processBatch(batch, id)
			}
			log.Printf("Worker %d stopped", id)
			return

		case event := <-w.eventChannel:
			if event == nil {
				continue
			}

			batch = append(batch, event)

			// Flush khi batch đầy
			if len(batch) >= w.batchSize {
				w.processBatch(batch, id)
				batch = make([]*models.ClickEvent, 0, w.batchSize)
			}

		case <-ticker.C:
			// Flush theo interval
			if len(batch) > 0 {
				w.processBatch(batch, id)
				batch = make([]*models.ClickEvent, 0, w.batchSize)
			}
		}
	}
}

// processBatch xử lý một batch events
func (w *ClickAnalyticsWorker) processBatch(batch []*models.ClickEvent, workerID int) {
	if len(batch) == 0 {
		return
	}

	start := time.Now()
	successCount := 0
	errorCount := 0

	// Group events theo short_code để update click count hiệu quả
	clickCounts := make(map[string]int)

	for _, event := range batch {
		// Lưu click event vào database
		if err := w.analyticsRepo.SaveClickEvent(event); err != nil {
			log.Printf("Error saving click event: %v", err)
			errorCount++
			continue
		}

		successCount++
		clickCounts[event.ShortCode]++
	}

	// Batch update click counts
	for shortCode, count := range clickCounts {
		for i := 0; i < count; i++ {
			if err := w.urlRepo.IncrementClickCount(shortCode); err != nil {
				log.Printf("Error incrementing click count for %s: %v", shortCode, err)
			}
		}
	}

	elapsed := time.Since(start)
	log.Printf("Worker %d: Processed batch of %d events (%d success, %d errors) in %v",
		workerID, len(batch), successCount, errorCount, elapsed)
}

// GetQueueSize trả về số events đang chờ trong queue
func (w *ClickAnalyticsWorker) GetQueueSize() int {
	return len(w.eventChannel)
}

// GetStats trả về thống kê của worker
func (w *ClickAnalyticsWorker) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"queue_size":     w.GetQueueSize(),
		"worker_count":   w.workerCount,
		"batch_size":     w.batchSize,
		"flush_interval": w.flushInterval.String(),
		"is_running":     w.isRunning,
	}
}
