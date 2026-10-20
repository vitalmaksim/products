package supplier

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/time/rate"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	limiter    *rate.Limiter
}

func NewClient(client *http.Client, baseURL string, limiter *rate.Limiter) *Client {
	return &Client{httpClient: client, baseURL: baseURL, limiter: limiter}
}

func (c *Client) GetProducts(ctx context.Context, page int) (ProductsResponse, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return ProductsResponse{}, fmt.Errorf("parse supplier url failed: %w", err)
	}

	u = u.JoinPath("products")
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return ProductsResponse{}, fmt.Errorf("create supplier request: %w", err)
	}

	// Все воркеры ждут один лимитер; отмена синхронизации прерывает и это ожидание.
	if err := c.limiter.Wait(ctx); err != nil {
		return ProductsResponse{}, fmt.Errorf("wait supplier rate limit: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ProductsResponse{}, fmt.Errorf("unable to fetch request %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ProductsResponse{}, fmt.Errorf("invalid supplier api status")
	}

	var result ProductsResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return ProductsResponse{}, fmt.Errorf("unable to decode json %w", err)
	}

	return result, nil
}
