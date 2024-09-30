package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/nianticlabs/venator/internal/config"
	"github.com/nianticlabs/venator/internal/signal"
	"github.com/sirupsen/logrus"
)

var logger = logrus.StandardLogger().WithField("pkg", "connector/pubsub")

type Client struct {
	projectID string
	topicID   string
}

func New(ctx context.Context, config Config) (*Client, error) {
	return &Client{
		projectID: config.ProjectID,
		topicID:   config.TopicID,
	}, nil
}

func (c *Client) Publish(ctx context.Context, results []map[string]string, cfg *config.RuleConfig) error {
	if len(results) == 0 {
		return nil
	}

	client, err := pubsub.NewClient(ctx, c.projectID)
	if err != nil {
		return err
	}
	defer client.Close()

	topic := client.Topic(c.topicID)
	if ok, err := topic.Exists(ctx); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("topic %s doesn't exist", c.topicID)
	}

	pubResults := make(chan *pubsub.PublishResult)
	go func() {
		defer close(pubResults)
		for _, r := range results {
			msg, err := buildPubSubMessage(r, cfg)
			if err != nil {
				logger.Errorf("failed to build pubsub message: %v\n", err)
				pubResults <- nil
				continue
			}
			res := topic.Publish(ctx, msg)
			pubResults <- res
		}
	}()

	var pubErrors []error
	for res := range pubResults {
		if res == nil {
			continue
		}
		id, err := res.Get(ctx)
		if err != nil {
			logger.Errorf("failed to publish message %v: %v\n", id, err)
			pubErrors = append(pubErrors, err)
			continue
		}
		logger.Infof("published message with msg ID: %v\n", id)
	}
	if len(pubErrors) != 0 {
		return fmt.Errorf("get: %w", errors.Join(pubErrors...))
	}

	return nil
}

func buildPubSubMessage(result map[string]string, cfg *config.RuleConfig) (*pubsub.Message, error) {
	output, err := signal.BuildOutput(result, cfg)
	if err != nil {
		return nil, err
	}

	dataJSON, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}

	return &pubsub.Message{
		Data: dataJSON,
	}, nil
}
