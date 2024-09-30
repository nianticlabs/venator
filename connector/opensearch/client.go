package opensearch

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/opensearch-project/opensearch-go/v2"

	"github.com/nianticlabs/venator/internal/config"
	"github.com/nianticlabs/venator/internal/signal"
)

type Client struct {
	osConfig    Config
	osClient    *opensearch.Client
	osTransport *http.Transport
}

const (
	outputIndexName = "signals"
	sqlPluginPath   = "/_plugins/_sql"
)

func New(ctx context.Context, config Config) (*Client, error) {
	// Create opensearch client
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: config.InsecureSkipVerify}, // #nosec G402
	}
	osc, err := opensearch.NewClient(
		opensearch.Config{
			Addresses: []string{config.URL},
			Username:  config.Username,
			Password:  config.Password,
			Transport: customTransport,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create opensearch client: %w", err)
	}

	return &Client{
		osClient:    osc,
		osConfig:    config,
		osTransport: customTransport,
	}, nil
}

func (c *Client) Query(ctx context.Context, cfg *config.RuleConfig) ([]map[string]string, error) {
	body, err := json.Marshal(map[string]string{"query": cfg.Query})
	if err != nil {
		return nil, err
	}

	// Create an HTTP client required for calling the SQL/PPL API (not supported in the opensearch-go Client).
	httpClient := &http.Client{Transport: c.osTransport}
	req, err := http.NewRequest("POST", c.osConfig.URL+sqlPluginPath, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.osConfig.Username, c.osConfig.Password)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("server responded with unexpected status code %d: %w", resp.StatusCode, err)
		}
		var queryError QueryErrorResponse
		if err = json.Unmarshal(errBody, &queryError); err != nil {
			return nil, fmt.Errorf("server responded with unexpected status code %d: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("server responded with unexpected status code %d: %s (details: %s)",
			resp.StatusCode, queryError.Error.Reason, queryError.Error.Details)
	}

	var response QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding QueryResponse: %w", err)
	}
	var results []map[string]string
	for _, row := range response.Datarows {
		res := make(map[string]string)
		for i, col := range response.Schema {
			value := fmt.Sprintf("%v", row[i])
			if _, exists := col["alias"]; exists {
				res[col["alias"]] = value
			} else {
				res[col["name"]] = value
			}
		}
		results = append(results, res)
	}

	return results, nil
}

func (c *Client) Publish(ctx context.Context, results []map[string]string, cfg *config.RuleConfig) error {
	if len(results) == 0 {
		return nil
	}

	body, err := buildBulkRequestBody(results, cfg)
	if err != nil {
		return err
	}

	bulkResp, err := c.osClient.Bulk(strings.NewReader(body))
	if err != nil {
		return err
	}
	var bulkResponse BulkQueryResponse
	if err := json.NewDecoder(bulkResp.Body).Decode(&bulkResponse); err != nil {
		return fmt.Errorf("error decoding BulkQueryResponse: %w", err)
	}

	var errArr []error
	if bulkResponse.Errors {
		for _, item := range bulkResponse.Items {
			for action, result := range item {
				if status, ok := result["status"].(float64); ok && status >= 400 {
					errArr = append(errArr, fmt.Errorf("error in %s: %d", action, int(status)))
				}
			}
		}
	}
	return errors.Join(errArr...)
}

func buildBulkRequestBody(results []map[string]string, cfg *config.RuleConfig) (string, error) {
	var body strings.Builder

	for _, r := range results {
		output, err := signal.BuildOutput(r, cfg)
		if err != nil {
			return "", err
		}

		docJSON, err := json.Marshal(output)
		if err != nil {
			return "", err
		}

		bReq := BulkRequestOp{
			Create: &CreateReq{
				Index: outputIndexName,
			},
		}
		jsonBytes, err := json.Marshal(bReq)
		if err != nil {
			return "", err
		}
		actionLine := string(jsonBytes)
		body.WriteString(actionLine + "\n")
		body.Write(docJSON)
		body.WriteString("\n")
	}

	return body.String(), nil
}
