package main

import (
	"context"

	"github.com/alexflint/go-arg"
	"github.com/sirupsen/logrus"

	"github.com/nianticlabs/venator/connector"
	"github.com/nianticlabs/venator/internal/config"
	"github.com/nianticlabs/venator/internal/exclusion"
	"github.com/nianticlabs/venator/internal/llm"
	llmconfig "github.com/nianticlabs/venator/internal/llm/config"
	"github.com/nianticlabs/venator/internal/llm/model"
)

var logger = logrus.StandardLogger()

var args struct {
	RuleConfigPath   string `arg:"-r,--rule-config,required" help:"Path to the rule configuration file"`
	GlobalConfigPath string `arg:"-c,--global-config" help:"Path to the global configuration file" default:"config/files/global_config.yaml"`
	LogLevel         string `arg:"-l,--log-level" help:"Log level" default:"info"`
}

func main() {
	ctx := context.Background()
	arg.MustParse(&args)
	setLogLevel(args.LogLevel)

	ruleCfg, err := config.ParseRuleConfig(args.RuleConfigPath)
	if err != nil {
		logger.Fatalf("error reading rule config: %s", err)
	}

	globalCfg, err := config.ParseGlobalConfig(args.GlobalConfigPath)
	if err != nil {
		logger.Fatalf("error reading global config: %s", err)
	}

	connectorRegistry := connector.NewRegistry(ctx, globalCfg)

	qr, err := connectorRegistry.GetQueryRunner(ruleCfg.QueryEngine)
	if err != nil {
		logger.Fatalf("error retrieving query runner '%s': %s", ruleCfg.QueryEngine, err)
	}

	var publishers []connector.Publisher
	for _, pubName := range ruleCfg.Publishers {
		pub, err := connectorRegistry.GetPublisher(pubName)
		if err != nil {
			logger.Fatalf("error retrieving publisher '%s': %s", pubName, err)
		}
		publishers = append(publishers, pub)
	}

	var excluder *exclusion.Excluder
	if ruleCfg.ExclusionsPath != "" {
		excluder, err = exclusion.NewExcluder(ruleCfg.ExclusionsPath)
		if err != nil {
			logger.Fatalf("error initializing exclusions: %s", err)
		}
		logger.Infof("Loaded exclusions from %s", ruleCfg.ExclusionsPath)
	}

	var llmClient model.Client
	if ruleCfg.LLM != nil && ruleCfg.LLM.Enabled {
		llmConfig := llmconfig.Config{
			Provider:    llmconfig.Provider(globalCfg.LLM.Provider),
			APIKey:      globalCfg.LLM.APIKey,
			Model:       globalCfg.LLM.Model,
			ServerURL:   globalCfg.LLM.ServerURL,
			Temperature: globalCfg.LLM.Temperature,
		}

		llmClient, err = llm.New(llmConfig)
		if err != nil {
			logger.Fatalf("error initializing LLM: %s", err)
		}
	}

	parsedResponse, err := qr.Query(ctx, ruleCfg)
	if err != nil {
		logger.Fatalf("error running the query: %s", err)
	}

	if excluder != nil {
		var filtered []map[string]string
		for _, result := range parsedResponse {
			if excluder.IsExcluded(result) {
				logger.Debugf("Excluded result: %+v", result)
				continue
			}
			filtered = append(filtered, result)
		}
		parsedResponse = filtered
		logger.Infof("After exclusions, %d results remain", len(parsedResponse))
	}

	if ruleCfg.LLM != nil && ruleCfg.LLM.Enabled {
		parsedResponse, err = llm.Process(ctx, llmClient, parsedResponse, ruleCfg)
		if err != nil {
			logger.Errorf("error processing LLM: %s", err)
			return
		}
		if len(parsedResponse) == 0 {
			logger.Infof("No results from LLM to publish")
			return
		}
		logger.Infof("LLM processing completed successfully")
	}

	if len(parsedResponse) == 0 {
		logger.Infof("No results to publish")
		return
	}

	for i, pub := range publishers {
		if err := pub.Publish(ctx, parsedResponse, ruleCfg); err != nil {
			logger.Errorf("error publishing to '%s': %s", ruleCfg.Publishers[i], err)
		} else {
			logger.Infof("Successfully published to '%s'", ruleCfg.Publishers[i])
		}
	}
}

func setLogLevel(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logger.Fatalf("error parsing log level: %s", err)
	}
	logger.SetLevel(lvl)
}
