package bigquery

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/bigquery"
	"github.com/nianticlabs/venator/internal/config"
	"github.com/nianticlabs/venator/internal/signal"
	"google.golang.org/api/iterator"
)

type Client struct {
	client   *bigquery.Client
	inserter *bigquery.Inserter
}

func New(ctx context.Context, config Config) (*Client, error) {
	var inserter *bigquery.Inserter
	client, err := bigquery.NewClient(ctx, config.ProjectID)
	if err != nil {
		return nil, err
	}
	if config.DatasetID != "" && config.TableID != "" {
		// Check if the dataset and table exist
		if _, err := client.Dataset(config.DatasetID).Metadata(ctx); err != nil {
			return nil, err
		}
		if _, err := client.Dataset(config.DatasetID).Table(config.TableID).Metadata(ctx); err != nil {
			return nil, err
		}
		inserter = client.Dataset(config.DatasetID).Table(config.TableID).Inserter()
	}

	return &Client{
		client:   client,
		inserter: inserter,
	}, nil
}

func (c *Client) Query(ctx context.Context, cfg *config.RuleConfig) ([]map[string]string, error) {
	query := c.client.Query(cfg.Query)
	it, err := query.Read(ctx)
	if err != nil {
		return nil, err
	}

	var results []map[string]string
	for {
		var row map[string]bigquery.Value
		err := it.Next(&row)
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		res := make(map[string]string)
		for k, v := range row {
			res[k] = fmt.Sprintf("%v", v)
		}
		results = append(results, res)
	}

	return results, nil
}

func (c *Client) Publish(ctx context.Context, results []map[string]string, cfg *config.RuleConfig) error {
	if len(results) == 0 {
		return nil
	}

	for _, r := range results {
		output, err := signal.BuildOutput(r, cfg)
		if err != nil {
			return err
		}

		if err := c.inserter.Put(ctx, output); err != nil {
			return err
		}
	}

	return nil
}
