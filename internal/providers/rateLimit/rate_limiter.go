package base

import (
	"context"

	"github.com/Achno/gowall/internal/providers"
	"golang.org/x/time/rate"
)

// implements OCRProvider interface
type RateLimitedProvider struct {
	provider providers.OCRProvider
	limiter  *rate.Limiter
}

// WithRateLimit wraps an OCRProvider to rate limit its OCR calls. If rps <= 0 rate limiting is disabled.
func WithRateLimit(provider providers.OCRProvider, rps float64, burst int) providers.OCRProvider {
	if rps <= 0 {
		return provider
	}

	return &RateLimitedProvider{
		provider: provider,
		limiter:  rate.NewLimiter(rate.Limit(rps), burst),
	}
}

func (r *RateLimitedProvider) OCR(ctx context.Context, input providers.OCRInput) (*providers.OCRResult, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return r.provider.OCR(ctx, input)
}

func (r *RateLimitedProvider) OCRBatch(ctx context.Context, inputs []providers.OCRInput) ([]*providers.OCRResult, error) {
	return providers.ProcessBatchWithPDFFallback(
		ctx,
		r.provider,
		r.provider.OCR,
		inputs,
		"provider",
		r.limiter,
	)
}

// Implements "PDFCapable" interface and return the result of the wrapped provider otherwise false
func (r *RateLimitedProvider) SupportsPDF() bool {
	if pdfCapable, ok := r.provider.(providers.PDFCapable); ok {
		return pdfCapable.SupportsPDF()
	}
	return false
}

// Implements the "RateLimited" interface
func (r *RateLimitedProvider) SetRateLimit(rps float64, burst int) {
	r.limiter = rate.NewLimiter(rate.Limit(rps), burst)
}
