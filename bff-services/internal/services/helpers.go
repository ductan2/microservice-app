package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"bff-services/internal/types"
)

// doRequest performs HTTP requests for service clients
func doRequest(ctx context.Context, baseURL, method, path string, httpClient *http.Client, payload interface{}, headers http.Header) (*types.HTTPResponse, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("service base URL is not configured")
	}

	endpoint := baseURL + path

	var bodyReader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &types.HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header.Clone(),
	}, nil
}

// internalAuthHeaders creates headers for internal microservice communication
func internalAuthHeaders(userID, email, sessionID string) http.Header {
	header := http.Header{}
	header.Set("X-User-ID", userID)
	header.Set("X-User-Email", email)
	header.Set("X-Session-ID", sessionID)
	return header
}

// bearerAuthHeader builds the Authorization header for downstream services expecting JWT tokens.
func bearerAuthHeader(token string) http.Header {
	header := http.Header{}
	if token != "" {
		header.Set("Authorization", "Bearer "+token)
	}
	return header
}
