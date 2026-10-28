package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	productModel "order-management/product/model"

	"github.com/sony/gobreaker"
)

// Service matches the contract expected by both OrderService and GraphQL
type Service interface {
	GetProductPrice(ctx context.Context, productID uint) (float64, error)
	GetProductDetails(ctx context.Context, productID uint) (*productModel.Product, error)
}

type remoteProductService struct {
	baseURL    string
	httpClient *http.Client
	cb         *gobreaker.CircuitBreaker
}

func NewRemoteProductService(baseURL string) Service {
	cbSettings := gobreaker.Settings{
		Name:        "ProductServiceCircuitBreaker",
		MaxRequests: 3,
		Interval:    10 * time.Second,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	}

	return &remoteProductService{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
		cb: gobreaker.NewCircuitBreaker(cbSettings),
	}
}

// GetProductPrice calls GetProductDetails and returns just the price
func (c *remoteProductService) GetProductPrice(ctx context.Context, productID uint) (float64, error) {
	prod, err := c.GetProductDetails(ctx, productID)
	if err != nil {
		return 0, err
	}
	return prod.Price, nil
}

// GetProductDetails fetches the full product via Circuit Breaker + Retries
func (c *remoteProductService) GetProductDetails(ctx context.Context, productID uint) (*productModel.Product, error) {
	result, err := c.cb.Execute(func() (interface{}, error) {
		return c.fetchWithRetry(ctx, productID, 3)
	})
	if err != nil {
		return nil, fmt.Errorf("circuit breaker rejected or call failed: %w", err)
	}

	prod, ok := result.(*productModel.Product)
	if !ok {
		return nil, errors.New("invalid product format received")
	}
	return prod, nil
}

func (c *remoteProductService) fetchWithRetry(ctx context.Context, productID uint, maxRetries int) (*productModel.Product, error) {
	var lastErr error
	url := fmt.Sprintf("%s/api/v1/products/%d", c.baseURL, productID)

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Context cancellation check
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Exponential Backoff
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * 100 * time.Millisecond
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return nil, errors.New("product not found")
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("unexpected status code: %d", resp.StatusCode)
			continue
		}

		var prod productModel.Product
		if err := json.NewDecoder(resp.Body).Decode(&prod); err != nil {
			resp.Body.Close()
			lastErr = err
			continue
		}
		resp.Body.Close()

		return &prod, nil
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}