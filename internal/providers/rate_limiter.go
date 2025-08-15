package providers

import (
	"context"
	"fmt"

	"golang.org/x/time/rate"
)

// implements OCRProvider interface
type RateLimitedProvider struct {
	provider OCRProvider
	limiter  *rate.Limiter
}

// WithRateLimit wraps an OCRProvider to rate limit its OCR calls. If rps <= 0 rate limiting is disabled.
func WithRateLimit(provider OCRProvider, rps float64, burst int) OCRProvider {
	if rps <= 0 {
		return provider
	}

	return &RateLimitedProvider{
		provider: provider,
		limiter:  rate.NewLimiter(rate.Limit(rps), burst),
	}
}

func (r *RateLimitedProvider) OCR(ctx context.Context, input OCRInput) (*OCRResult, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return r.provider.OCR(ctx, input)
}

// Complete implements TextProcessorProvider interface - delegates to the wrapped provider if it supports text processing
func (r *RateLimitedProvider) Complete(ctx context.Context, text string) (string, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return "", err
	}

	if textProcessor, ok := r.provider.(TextProcessorProvider); ok {
		return textProcessor.Complete(ctx, text)
	}

	return "", fmt.Errorf("wrapped provider does not implement TextProcessorProvider interface")
}

func (r *RateLimitedProvider) GetConfig() Config {
	return r.provider.GetConfig()
}

// Implements "PDFCapable" interface and return the result of the wrapped provider otherwise false
func (r *RateLimitedProvider) SupportsPDF() bool {
	if pdfCapable, ok := r.provider.(PDFCapable); ok {
		return pdfCapable.SupportsPDF()
	}
	return false
}

// Implements the "RateLimited" interface
func (r *RateLimitedProvider) SetRateLimit(rps float64, burst int) {
	r.limiter = rate.NewLimiter(rate.Limit(rps), burst)
}

func (r *RateLimitedProvider) GetRateLimiter() *rate.Limiter {
	return r.limiter
}
