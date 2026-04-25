package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"
)

const TrueUA = "frozeel-bot/0.1 (https://github.com/21ess/frozeel)"

var defaultHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

// DoHTTPJSON help to send request and parse response, the `result` is required a pointer type
func DoHTTPJSON(ctx context.Context, method, url string, body []byte, token string, result any) error {
	var reqBody *bytes.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("build request failed: %w", err)
	}

	req.Header.Set("User-Agent", TrueUA)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := defaultHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	// result requires to be a pointer type
	if reflect.TypeOf(result).Kind() != reflect.Pointer {
		return fmt.Errorf("result must be a pointer")
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
