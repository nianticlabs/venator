// To run the integration tests locally, you need to run docker-compose from the current directory.
// ```bash
// docker-compose up -d
// docker-compose ps
// go test -v ./
// docker-compose stop
// ```

package opensearch_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nianticlabs/venator/connector/opensearch"
	"github.com/nianticlabs/venator/internal/config"
)

const (
	url      = "https://localhost:9200"
	username = "admin"
	password = "dummy-password-9QW7dB4@HD"
	index    = "opensearch_dashboards_sample_data_logs"
)

var client *opensearch.Client

func TestQueryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	if client == nil {
		initTestEnvironment()
	}

	tests := []struct {
		name       string
		cfg        *config.RuleConfig
		errMessage string
		expected   []map[string]string
	}{
		{
			name:     "successful query with non-empty result",
			cfg:      &config.RuleConfig{Query: fmt.Sprintf("SELECT timestamp,ip,host,request FROM %s LIMIT 2", index)},
			expected: []map[string]string{{}, {}},
		},
		{
			name:       "wrong sql query",
			cfg:        &config.RuleConfig{Query: "SELECT FROM"},
			errMessage: "Invalid SQL query",
		},
	}

	for _, tt := range tests {
		ctx := context.Background()
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.RuleConfig{}
			*cfg = *tt.cfg
			result, err := client.Query(ctx, cfg)
			if tt.errMessage == "" {
				if err != nil {
					t.Fatalf("did not expect an error but got: %v", err)
				}
				if diff := cmp.Diff(len(tt.expected), len(result)); diff != "" {
					t.Fatalf("unexpected result (-want +got):\n%s", diff)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected an error but got nil")
			}
			if !strings.Contains(err.Error(), tt.errMessage) {
				t.Fatalf("expected error message to contain %q, got %q", tt.errMessage, err.Error())
			}
		})
	}
}

func TestPublishIntegration(t *testing.T) {
	ctx := context.Background()
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if client == nil {
		initTestEnvironment()
	}

	tests := []struct {
		name string
		data []map[string]string
		err  error
	}{
		{
			name: "successful publish with non-empty data",
			data: []map[string]string{
				{"timestamp": "2021-09-01T00:00:00Z", "message": "raw data 1", "host": "resource 1", "actor": "actor 1"},
				{"timestamp": "2021-09-02T00:00:00Z", "message": "raw data 2", "host": "resource 2", "actor": "actor 2"},
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.RuleConfig{
				Name: "testSignal",
				UID:  "testUID",
				Output: config.Output{
					Format: "signal",
					Fields: []config.OutputField{
						{Field: "Timestamp", Source: "timestamp"},
						{Field: "Message", Source: "message"},
						{Field: "ResourceName", Source: "host"},
						{Field: "ActorUserName", Source: "actor"},
					},
				},
			}

			err := client.Publish(ctx, tt.data, cfg)
			if !errors.Is(err, tt.err) {
				t.Fatalf("expected error: %v, got: %v", tt.err, err)
			}
		})
	}
}

func initTestEnvironment() {
	cfg := opensearch.Config{
		URL:                url,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: true,
	}
	var err error
	ctx := context.Background()
	client, err = opensearch.New(ctx, cfg)
	if err != nil {
		fmt.Printf("unable to initialize OpenSearch client: %v\n", err)
		os.Exit(1)
	}

	if err := addSampleData(); err != nil {
		fmt.Printf("unable to add sample data: %v\n", err)
		os.Exit(1)
	}
}

func addSampleData() error {
	url := "http://localhost:5601/api/sample_data/logs"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("osd-xsrf", "true")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("server responded with unexpected status code %d", resp.StatusCode)
	}

	return nil
}
