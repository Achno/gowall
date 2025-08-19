package providers

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Achno/gowall/utils"
)

// 1. You create a custom progress tracker, start it pass it as an argument to the stages, have the increment/decrement there
// progress := WithCustomProgress(len(initialItems), func(inProgress, completed, failed, total int64) string {
// 	return fmt.Sprintf("Pre-processing: %d active, %d done, %d failed (Total: %d)",
// 		inProgress, completed, failed, total)
// })
// progress.Start()
// defer progress.Stop()

// grayScaleStage := NewGrayScaleStageWithProgress(image.GrayScaleProcessor{}, progress)

// ProgressTracker handles progress tracking with atomic counters and periodic updates
type ProgressTracker struct {
	total      int64
	completed  atomic.Int64
	failed     atomic.Int64
	inProgress atomic.Int64

	ticker     *time.Ticker
	ctx        context.Context
	cancel     context.CancelFunc
	updateFreq time.Duration

	messageFormatter func(inProgress, completed, failed, total int64) string
	prefix           string
	useCustomFormat  bool
}

type ProgressTrackerConfig struct {
	Total            int64
	UpdateFrequency  time.Duration
	MessageFormatter func(inProgress, completed, failed, total int64) string
	Prefix           string
}

// NewProgressTracker creates a new progress tracker
func NewProgressTracker(config ProgressTrackerConfig) *ProgressTracker {
	updateFreq := config.UpdateFrequency
	if updateFreq == 0 {
		updateFreq = 500 * time.Millisecond
	}

	formatter := config.MessageFormatter
	prefix := config.Prefix
	useCustomFormat := formatter != nil

	if formatter == nil {
		formatter = DefaultMessageFormatter(prefix)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ProgressTracker{
		total:            config.Total,
		updateFreq:       updateFreq,
		messageFormatter: formatter,
		prefix:           prefix,
		useCustomFormat:  useCustomFormat,
		ctx:              ctx,
		cancel:           cancel,
	}
}

// Start begins the progress tracking with periodic updates
func (pt *ProgressTracker) Start() {
	if pt.total == 0 {
		return
	}

	// Initialize all items as "in progress"
	pt.inProgress.Store(pt.total)

	pt.ticker = time.NewTicker(pt.updateFreq)
	utils.Spinner.Start()

	go func() {
		for {
			select {
			case <-pt.ctx.Done():
				return
			case <-pt.ticker.C:
				pt.updateMessage()
			}
		}
	}()
}

func (pt *ProgressTracker) updateMessage() {
	inProgress, completed, failed, total := pt.GetCounters()
	message := pt.messageFormatter(inProgress, completed, failed, total)
	utils.Spinner.Message(message)
}

func (pt *ProgressTracker) Stop(message string) {
	if pt.cancel != nil {
		pt.cancel()
	}
	if pt.ticker != nil {
		pt.ticker.Stop()
	}
	utils.Spinner.StopMessage(message)
	utils.Spinner.Stop()
}

func (pt *ProgressTracker) IncrementCompleted() {
	pt.completed.Add(1)
	pt.inProgress.Add(-1)
}

func (pt *ProgressTracker) IncrementFailed() {
	pt.failed.Add(1)
	pt.inProgress.Add(-1)
}

func (pt *ProgressTracker) IncrementInProgress() {
	pt.inProgress.Add(1)
}

func (pt *ProgressTracker) GetCounters() (inProgress, completed, failed, total int64) {
	return pt.inProgress.Load(), pt.completed.Load(), pt.failed.Load(), pt.total
}

func (pt *ProgressTracker) IsComplete() bool {
	return pt.completed.Load()+pt.failed.Load() == pt.total
}

func (pt *ProgressTracker) SetTotal(newTotal int64) {
	pt.total = newTotal

	// Adjust inProgress count to match the new total
	// (completed + failed should remain the same)
	completed := pt.completed.Load()
	failed := pt.failed.Load()
	newInProgress := newTotal - completed - failed

	if newInProgress >= 0 {
		pt.inProgress.Store(newInProgress)
	}
}

// SetPrefix updates the prefix for the progress tracker (only works with default formatter)
func (pt *ProgressTracker) SetPrefix(newPrefix string) {
	if !pt.useCustomFormat {
		pt.prefix = newPrefix
		pt.messageFormatter = DefaultMessageFormatter(pt.prefix)
	}
}

// AppendToPrefix concatenates a string to the current prefix (only works with default formatter)
func (pt *ProgressTracker) AppendToPrefix(suffix string) {
	if !pt.useCustomFormat {
		pt.prefix = pt.prefix + suffix
		pt.messageFormatter = DefaultMessageFormatter(pt.prefix)
	}
}

// GetPrefix returns the current prefix (only meaningful with default formatter)
func (pt *ProgressTracker) GetPrefix() string {
	return pt.prefix
}

// DefaultMessageFormatter creates a default message formatter with the given prefix
func DefaultMessageFormatter(prefix string) func(inProgress, completed, failed, total int64) string {
	if prefix == "" {
		prefix = "Progress"
	}
	return func(inProgress, completed, failed, total int64) string {
		return fmt.Sprintf("%s: %d processing, %d completed, %d failed (Total: %d)",
			prefix, inProgress, completed, failed, total)
	}
}

// WithGenericProgress creates a progress tracker with default settings
func WithGenericProgress(total int) *ProgressTracker {
	return NewProgressTracker(ProgressTrackerConfig{
		Total: int64(total),
	})
}

// WithCustomProgress creates a progress tracker with custom message formatter
func WithCustomProgress(total int, formatter func(inProgress, completed, failed, total int64) string) *ProgressTracker {
	return NewProgressTracker(ProgressTrackerConfig{
		Total:            int64(total),
		MessageFormatter: formatter,
	})
}

// WithPrefixProgress creates a progress tracker with a custom prefix
func WithPrefixProgress(total int, prefix string) *ProgressTracker {
	return NewProgressTracker(ProgressTrackerConfig{
		Total:  int64(total),
		Prefix: prefix,
	})
}
