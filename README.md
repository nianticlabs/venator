<p align="center">
<img src="docs/images/logo.png" width="200"/>
</p>

# Venator - Threat Detection Platform

**A flexible detection system that simplifies rule management and deployment with K8s CronJob and Helm.**

Venator is optimized for Kubernetes deployment but is flexible enough to run standalone or with other job schedulers like Nomad. It provides a highly adaptable detection engine that prioritizes simplicity, extensibility, and ease of maintenance. Supporting multiple query engines and publishers, Venator allows you to easily switch between different data lakes or services with minimal changes, avoiding vendor lock-in and dependence on specific SIEM solutions for signal generation.

## Why Venator?

Many existing open-source and commercial threat detection solutions lack effective tools for monitoring and managing scheduled detection rules. Common challenges include verifying whether detection jobs ran successfully, troubleshooting failed jobs, and running backfills or ad-hoc executions. Moreover, adding new detection rules or integrating new log sources often leads to unnecessary complexity.

## How It Works

Venator operates by running each detection rule as an independent job, allowing for flexible query execution and result handling. Each rule uses a query engine (e.g., OpenSearch, BigQuery) to fetch data, process the results, and publish the findings to one or more destinations like BigQuery for signal storage, or PubSub for alerts to trigger your automation system. This modular approach ensures that the failure of one rule doesn’t impact others.

### Key Components:

- **Detection Rules**: Detection logic is defined in YAML files. Each rule specifies its own query engine and publishers, making it possible to query different data lakes in parallel or deliver results to different platforms. For example, one rule could query OpenSearch logs and publish alerts to PubSub, while another queries BigQuery and sends results to Slack. You can see some example rules [here](config/rules/).
  
- **Job Execution**: Venator schedules and runs each rule as a separate Kubernetes CronJob (or another job scheduler like Nomad). This scheduling allows rules to run at regular intervals (e.g., hourly, daily) or on-demand for ad-hoc queries. Kubernetes handles the lifecycle of these jobs, ensuring each rule runs in isolation.

- **Exclusions**: To reduce false positives, rules can reference exclusion lists, which filter out known benign events from the results before they’re published. These exclusion lists are also defined in YAML and support `and` and `or` conditions with operators like `equals`, `not_equals`, `contains`, `regex`, `in`, and `not_in`. Here is an [example](config/exclusions/example-rule.yaml) exclusion list.

- **LLM Integration**: Venator integrates with Large Language Models (LLMs) to provide enhanced signal analysis. This is particularly useful for analyzing or correlating lower-confidence signals that may not be suitable for immediate alerts.

- **Automated Deployment**: Venator's deployment model uses Helm to automate the process. Helm charts manage configuration files like detection rules, exclusions, and global settings as Kubernetes ConfigMaps. Through a CI/CD pipeline, any changes to detection rules or code automatically trigger new deployments, ensuring the system is always up-to-date without manual intervention.

## Deployment Guide

For detailed steps on deploying Venator using Helm and Kubernetes, see the [Deployment Guide](docs/deployment.md).