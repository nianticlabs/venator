<p align="center">
<img src="docs/images/logo.png" width="200"/>
</p>

# Venator - Threat Detection Platform

**A cloud-ready, customizable threat detection system with support for scheduled, ad-hoc, and multi-stage detections.**

Venator is a flexible threat detection platform designed to provide full control over the execution, monitoring, and management of detection rules. By leveraging existing technologies like search engines (OpenSearch, BigQuery) and job schedulers (Kubernetes CronJob, HashiCorp Nomad), Venator offers a highly adaptable detection engine that focuses on simplicity, extensibility, and ease of maintenance.

## Why?

Many existing open-source and commercial threat detection solutions lack the ability to reliably monitor and manage scheduled detection rules. Key limitations include the difficulty of ensuring whether detection jobs ran successfully, the inability to troubleshoot failed jobs, and challenges in running backfills or ad-hoc executions. Moreover, adding new detection rules or supporting additional log sources often leads to unnecessary complexity.

Venator was designed to address these gaps by leveraging existing infrastructure for job scheduling and query execution, while also offering a "Detection-as-Code" approach. This allows users to define detection rules as version-controlled YAML files, simplifying the process of rule creation, management, and deployment. Venator provides a lightweight, easy-to-maintain alternative to traditional SIEM detection engines, focusing on simplicity and flexibility without unnecessary complexity.

### Key Features:

- **Scheduled & Ad-hoc Detections**: Run rules on a scheduled basis or execute them retroactively for historical analysis.
- **Customizable Detection Languages**: Write detection logic using SQL, PPL, PQL, or other query languages depending on the underlying engine.
- **Full Monitoring & Control**: Ensure detection jobs run successfully, with built-in support for monitoring job execution, failures, and troubleshooting.
- **Multi-Stage Detections**: Enable signal correlation and multi-stage detections, offering flexibility for advanced threat detection patterns.
- **LLM Integration**: Venator can integrate with LLMs for enhanced signal analysis and correlation. This is particularly useful for low to medium fidelity signals that are insufficient for direct alerts.
- **Cloud-Ready**: Designed to be deployed on any cloud or on-prem infrastructure, leveraging Kubernetes, Helm, and GitLab CI for deployment automation.
- **Detection-as-Code**: Define, store, and version detection rules in YAML files, making it easy to track changes, automate deployments, and quickly iterate on detection logic.

## Deployment Guide

For detailed steps on how to deploy Venator using Helm and Kubernetes, including building Docker images, configuring connectors, and setting up detection rules, please refer to the [Deployment Guide](docs/deployment.md).
