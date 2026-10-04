package mercadopago

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// SearchParams contains optional query parameters for filtering payments.
type SearchParams struct {
	Sort      string
	Criteria  string
	BeginDate string
	EndDate   string
	Offset    int
	Limit     int
}

// Client defines the interface for communicating with MercadoPago API.
type Client interface {
	SearchPayments(ctx context.Context, params SearchParams) (*PaymentSearchResponse, error)
}

type client struct {
	baseURL     string
	accessToken string
	httpClient  *http.Client
}

// NewClient creates an instance of MercadoPago client.
func NewClient(baseURL, accessToken string) Client {
	return &client{
		baseURL:     baseURL,
		accessToken: accessToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchPayments calls the /v1/payments/search endpoint and parses the response into PaymentSearchResponse.
func (c *client) SearchPayments(ctx context.Context, params SearchParams) (*PaymentSearchResponse, error) {
	reqURL, err := url.Parse(fmt.Sprintf("%s/v1/payments/search", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	q := reqURL.Query()
	if params.Sort != "" {
		q.Set("sort", params.Sort)
	} else {
		q.Set("sort", "date_created")
	}

	if params.Criteria != "" {
		q.Set("criteria", params.Criteria)
	} else {
		q.Set("criteria", "desc")
	}

	if params.Limit > 0 {
		q.Set("limit", strconv.Itoa(params.Limit))
	} else {
		q.Set("limit", "50")
	}

	if params.Offset > 0 {
		q.Set("offset", strconv.Itoa(params.Offset))
	}

	if params.BeginDate != "" {
		q.Set("begin_date", params.BeginDate)
	}

	if params.EndDate != "" {
		q.Set("end_date", params.EndDate)
	}

	reqURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request execution failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("mercadopago api error: status=%d body=%s", resp.StatusCode, string(bodyBytes))
	}

	var searchResp PaymentSearchResponse
	if err := json.Unmarshal(bodyBytes, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response json: %w", err)
	}

	return &searchResp, nil
}
