package tezos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client is a basic TzKT API client.
type Client struct {
	// HTTP client
	client http.Client
	// Parsed URL for current protocol endpoint
	protoURL url.URL
	// Parsed URL for operation delegations endpoint
	delegURL url.URL
}

type Delegation struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Block     string    `json:"block"`
	Sender    struct {
		Address string `json:"address"`
	} `json:"sender"`
	Level  int32 `json:"level"`
	Amount int64 `json:"amount"`
}

// NewClient creates a new client and returns an error if the base URL passed is invalid.
func NewClient(baseURL string) (Client, error) {
	pBase, err := url.Parse(baseURL + "v1/protocols/current")
	if err != nil {
		return Client{}, err
	}

	dBase, err := url.Parse(baseURL + "v1/operations/delegations")
	if err != nil {
		return Client{}, err
	}

	return Client{
		protoURL: *pBase,
		delegURL: *dBase,
	}, nil
}

// GetCurrentProtocolTimeBetweenBlocks calls the "/protocols/current" endpoint of the
// TzKT API and returns the "timeBetweenBlocks" constant. Returns the underlying HTTP
// client errors, or any issues related to response processing.
func (c Client) GetCurrentProtocolTimeBetweenBlocks(ctx context.Context) (time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.protoURL.String(), nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad HTTP status: %d", resp.StatusCode)
	}

	payload := struct {
		Constants struct {
			TimeBetweenBlocks int32 `json:"timeBetweenBlocks"`
		} `json:"constants"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, err
	}

	return time.Duration(payload.Constants.TimeBetweenBlocks) * time.Second, nil
}

// GetDelegationsSince calls the "/operations/delegations" endpoint of the TzKT API
// and returns delegations which timestamps are greater or equal to the time passed
// as parameter. Delegation operations are guaranteed to be sorted most recent first
// by the TzKT API but this function does not enforce this behaviour.
// Returns the underlying HTTP client errors, or any issues related to response processing.
func (c Client) GetDelegationsSince(ctx context.Context, since time.Time) ([]Delegation, error) {
	// cannot sort by timestamp; sort by id since it seems to be a reliable increment
	url := c.delegURL.String() + "?select=id,sender,amount,level,timestamp,block&sort.asc=id&timestamp.ge=" + since.UTC().Format(time.RFC3339)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []Delegation{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return []Delegation{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []Delegation{}, fmt.Errorf("bad HTTP status: %d", resp.StatusCode)
	}

	payload := []Delegation{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return []Delegation{}, err
	}

	return payload, nil
}
