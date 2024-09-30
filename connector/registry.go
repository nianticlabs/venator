// registry.go
package connector

import (
	"context"
	"fmt"

	"github.com/nianticlabs/venator/connector/bigquery"
	"github.com/nianticlabs/venator/connector/opensearch"
	"github.com/nianticlabs/venator/connector/pubsub"
	"github.com/nianticlabs/venator/connector/slack"
	"github.com/nianticlabs/venator/internal/config"

	"github.com/sirupsen/logrus"
)

var logger = logrus.StandardLogger()

// Registry manages all connector instances for querying and publishing.
type Registry struct {
	queryRunners map[string]QueryRunner
	publishers   map[string]Publisher
}

// NewRegistry initializes the Registry with all configured connector instances.
func NewRegistry(ctx context.Context, globalCfg *config.GlobalConfig) *Registry {
	r := &Registry{
		queryRunners: make(map[string]QueryRunner),
		publishers:   make(map[string]Publisher),
	}

	r.initOpenSearch(ctx, globalCfg.OpenSearch)
	r.initPubSub(ctx, globalCfg.PubSub)
	r.initBigQuery(ctx, globalCfg.BigQuery)
	r.initSlack(ctx, globalCfg.Slack)

	return r
}

func (r *Registry) initOpenSearch(ctx context.Context, connectors config.OpenSearchConnectors) {
	if len(connectors.Instances) == 0 {
		logger.Debug("No OpenSearch instances configured. Skipping OpenSearch initialization.")
		return
	}

	for name, osCfg := range connectors.Instances {
		// Validate required fields
		if osCfg.URL == "" || osCfg.Username == "" || osCfg.Password == "" {
			logger.Warnf("Missing required fields for OpenSearch instance '%s'. Skipping initialization.", name)
			continue
		}

		client, err := opensearch.New(ctx, opensearch.Config{
			URL:                osCfg.URL,
			Username:           osCfg.Username,
			Password:           osCfg.Password,
			InsecureSkipVerify: osCfg.InsecureSkipVerify,
		})
		if err != nil {
			logger.Warnf("Error creating OpenSearch instance '%s': %v. Skipping initialization.", name, err)
			continue
		}

		instanceName := "opensearch." + name
		r.queryRunners[instanceName] = client
		r.publishers[instanceName] = client
		logger.Infof("Initialized OpenSearch instance '%s' as both QueryRunner and Publisher.", name)
	}
}

func (r *Registry) initPubSub(ctx context.Context, connectors config.PubSubConnectors) {
	if len(connectors.Instances) == 0 {
		logger.Debug("No PubSub instances configured. Skipping PubSub initialization.")
		return
	}

	for name, psCfg := range connectors.Instances {
		// Validate required fields
		if psCfg.ProjectID == "" || psCfg.TopicID == "" {
			logger.Warnf("Missing required fields for PubSub instance '%s'. Skipping initialization.", name)
			continue
		}

		client, err := pubsub.New(ctx, pubsub.Config{
			ProjectID: psCfg.ProjectID,
			TopicID:   psCfg.TopicID,
		})
		if err != nil {
			logger.Warnf("Error creating PubSub instance '%s': %v. Skipping initialization.", name, err)
			continue
		}

		instanceName := "pubsub." + name
		r.publishers[instanceName] = client
		logger.Infof("Initialized PubSub instance '%s' as Publisher.", name)
	}
}

func (r *Registry) initBigQuery(ctx context.Context, connectors config.BigQueryConnectors) {
	if len(connectors.Instances) == 0 {
		logger.Debug("No BigQuery instances configured. Skipping BigQuery initialization.")
		return
	}

	for name, bqCfg := range connectors.Instances {
		// Validate required fields
		if bqCfg.ProjectID == "" {
			logger.Warnf("Missing ProjectID for BigQuery instance '%s'. Skipping initialization.", name)
			continue
		}

		client, err := bigquery.New(ctx, bigquery.Config{
			ProjectID: bqCfg.ProjectID,
			DatasetID: bqCfg.DatasetID,
			TableID:   bqCfg.TableID,
		})
		if err != nil {
			logger.Warnf("Error creating BigQuery instance '%s': %v. Skipping initialization.", name, err)
			continue
		}

		instanceName := "bigquery." + name
		r.queryRunners[instanceName] = client
		r.publishers[instanceName] = client
		logger.Infof("Initialized BigQuery instance '%s' as both QueryRunner and Publisher.", name)
	}
}

func (r *Registry) initSlack(ctx context.Context, connectors config.SlackConnectors) {
	if len(connectors.Instances) == 0 {
		logger.Debug("No Slack instances configured. Skipping Slack initialization.")
		return
	}

	for name, slackCfg := range connectors.Instances {
		// Validate required fields
		if slackCfg.WebhookURL == "" {
			logger.Warnf("Missing webhookURL for Slack instance '%s'. Skipping initialization.", name)
			continue
		}

		client, err := slack.New(ctx, slack.Config{
			WebhookURL: slackCfg.WebhookURL,
		})
		if err != nil {
			logger.Warnf("Error creating Slack instance '%s': %v. Skipping initialization.", name, err)
			continue
		}

		instanceName := "slack." + name
		r.publishers[instanceName] = client
		logger.Infof("Initialized Slack instance '%s' as Publisher.", name)
	}
}

func (r *Registry) GetQueryRunner(name string) (QueryRunner, error) {
	qr, exists := r.queryRunners[name]
	if !exists {
		return nil, fmt.Errorf("query runner '%s' not found", name)
	}
	return qr, nil
}

func (r *Registry) GetPublisher(name string) (Publisher, error) {
	pub, exists := r.publishers[name]
	if !exists {
		return nil, fmt.Errorf("publisher '%s' not found", name)
	}
	return pub, nil
}
