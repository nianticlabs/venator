package connector

import (
	"context"

	"github.com/nianticlabs/venator/internal/config"
)

type Connector interface {
	New(ctx context.Context, config any) (Connector, error)
}

type QueryRunner interface {
	Query(ctx context.Context, ruleConfig *config.RuleConfig) ([]map[string]string, error)
}

type Publisher interface {
	Publish(ctx context.Context, data []map[string]string, ruleConfig *config.RuleConfig) error
}
