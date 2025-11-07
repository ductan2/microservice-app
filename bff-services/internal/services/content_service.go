package services

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"bff-services/internal/api/dto"
	"bff-services/internal/types"
	"github.com/redis/go-redis/v9"
)

// ContentService defines the contract for interacting with the content service GraphQL API.
type ContentService interface {
	ExecuteGraphQL(ctx context.Context, token string, payload dto.GraphQLRequest) (*types.HTTPResponse, error)
}

// ContentServiceClient implements ContentService against a remote HTTP GraphQL endpoint.
type ContentServiceClient struct {
	baseURL     string
	httpClient  *http.Client
	redisClient *redis.Client
}

// Whitelist of cacheable GraphQL operations with their TTL
var gqlCacheOps = map[string]time.Duration{
	"GetTopics": 5 * time.Minute,
	"GetLevels": 5 * time.Minute,
	"GetTags":   5 * time.Minute,
}

type graphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// NewContentServiceClient constructs a new ContentServiceClient.
func NewContentServiceClient(baseURL string, httpClient *http.Client) *ContentServiceClient {
	trimmed := strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &ContentServiceClient{
		baseURL:     trimmed,
		httpClient:  httpClient,
		redisClient: nil,
	}
}

// SetRedisClient sets the Redis client for caching
func (c *ContentServiceClient) SetRedisClient(redisClient *redis.Client) {
	c.redisClient = redisClient
}

// ExecuteGraphQL forwards raw GraphQL operations to the content service.
func (c *ContentServiceClient) ExecuteGraphQL(ctx context.Context, token string, payload dto.GraphQLRequest) (*types.HTTPResponse, error) {
	query := strings.TrimSpace(payload.Query)
	if query == "" {
		return nil, fmt.Errorf("graphql query is required")
	}

	request := graphQLRequest{Query: query}
	if len(payload.Variables) > 0 {
		request.Variables = payload.Variables
	}
	if op := strings.TrimSpace(payload.OperationName); op != "" {
		request.OperationName = op
	}

	// Check cache if Redis is configured and operation is whitelisted
	if c.redisClient != nil {
		opName := extractOperationName(query)
		if _, isCacheable := gqlCacheOps[opName]; isCacheable {
			cacheKey := generateCacheKey(opName, request.Variables)
			if cached, err := c.redisClient.Get(ctx, cacheKey).Result(); err == nil {
				log.Printf("Cache hit for operation: %s", opName)
				return &types.HTTPResponse{
					StatusCode: http.StatusOK,
					Body:       []byte(cached),
					Headers:    make(http.Header),
				}, nil
			}
		}
	}

	return c.sendGraphQLRequest(ctx, request, token)
}

func (c *ContentServiceClient) sendGraphQLRequest(ctx context.Context, payload graphQLRequest, token string) (*types.HTTPResponse, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("content service base URL is not configured")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal graphql payload: %w", err)
	}

	endpoint := c.baseURL + "/graphql"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create graphql request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform graphql request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read graphql response: %w", err)
	}

	httpResp := &types.HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header.Clone(),
	}

	// Cache successful responses for whitelisted operations
	if resp.StatusCode == http.StatusOK && c.redisClient != nil {
		opName := extractOperationName(payload.Query)
		if ttl, isCacheable := gqlCacheOps[opName]; isCacheable {
			cacheKey := generateCacheKey(opName, payload.Variables)
			// Cache asynchronously to avoid blocking response
			go func() {
				c.redisClient.Set(context.Background(), cacheKey, string(respBody), ttl)
				log.Printf("Cached response for operation: %s with TTL: %v", opName, ttl)
			}()
		}
	}

	return httpResp, nil
}

// extractOperationName extracts the operation name from a GraphQL query string
func extractOperationName(query string) string {
	re := regexp.MustCompile(`(?i)query\s+([A-Za-z0-9_]+)`)
	m := re.FindStringSubmatch(query)
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

// generateCacheKey creates a consistent cache key from operation name and variables
func generateCacheKey(opName string, variables map[string]interface{}) string {
	key := opName
	if len(variables) > 0 {
		varsJSON, _ := json.Marshal(variables)
		hash := md5.Sum(varsJSON)
		key = fmt.Sprintf("%s:%s", opName, hex.EncodeToString(hash[:]))
	}
	return fmt.Sprintf("gql:%s", key)
}
