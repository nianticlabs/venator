## Kubernetes/Helm Deployment Guide

This guide provides steps to deploy Venator on a Kubernetes cluster using Helm. Follow the instructions to set up the environment, build the Docker image, configure Helm, and deploy the application.

### Prerequisites

1. **Configure Cluster Access**  
   Ensure you have access to your Kubernetes cluster. Authenticate and get credentials for your cluster using the following command:

   ```bash
   gcloud container clusters get-credentials test-cluster --region=us-central1
   ```

   Then, check your current context and switch it if necessary:

   ```bash
   kubectl config get-contexts          # List available contexts
   kubectl config current-context       # Check current context
   kubectl config use-context <context-name>  # Switch to the correct context (replace <context-name> with your cluster's context)
   ```

2. **Create a Namespace (if not already created)**  
   To deploy Venator into its own namespace, create a new namespace:

   ```bash
   kubectl create namespace venator
   ```

### Build the Docker Image

To build the Venator Docker image, run the provided example script. This step assumes you're in the root directory containing the Dockerfile, and you're using an Artifact Registry repository named `venator-repo` to store built artifacts.

```bash
./scripts/build_image.sh
```

> **Note:** This is a manual build process. Ideally, it should be automated using a CI/CD pipeline like GitLab CI or GitHub Actions.

### Create Kubernetes Secrets

If you need to store credentials for your connectors, create Kubernetes secrets using the `scripts/create_secret.sh` or use any other secret manager you prefer. These secrets are referenced in `config/files/global_config.yaml` and `config/templates/cronjob.yaml`. Ensure they are configured correctly.

For example, set the OpenSearch password (if applicable):

```bash
read -s -r OPENSEARCH_PASSWORD
export OPENSEARCH_PASSWORD
```

In your `global_config.yaml`, reference the environment variable:

```yaml
opensearch:
  instances:
    prod:
      url: https://opensearch:9200
      username: user
      password: ${OPENSEARCH_PASSWORD}
      insecureSkipVerify: true
```

Ensure the corresponding environment variable is referenced correctly in `config/templates/cronjob.yaml`.

```yaml
...
    env:
    - name: OPENSEARCH_PASSWORD
        valueFrom:
        secretKeyRef:
            key: OSPASSWORD
            name: {{ $.Release.Name }}-secret
...
```

### Configure and Install the Helm Chart

1. **Modify the `values.yaml` File**  
   The Helm chart configuration is located in the `config` directory. Before deploying, modify the `values.yaml` file to reference your Docker image:

   ```yaml
   container:
     image: "your-docker-repo/venator-image:latest"  # Update with your Docker image
   ```

2. **Test Helm Chart Locally (Dry Run)**  
   Ensure everything is set up correctly by running a Helm template command to test rendering the chart templates:

   ```bash
   cd config
   helm template . --dry-run  # Test rendering chart templates locally
   ```

   You can also use the `--debug` flag for more detailed output:

   ```bash
   helm install venator-test . --dry-run --debug  # Render the chart locally without installing, checks for resource conflicts
   ```

3. **Install the Helm Chart**  
   Once everything looks good, install the Helm chart into the `venator` namespace:

   ```bash
   helm install venator-test . -n venator
   ```

4. **Upgrading the Chart**  
   To update the Helm chart, use the upgrade command:

   ```bash
   helm upgrade venator-test . -n venator
   ```

### Detection Rules

- Venator detection rules are defined in YAML files located in the `config/rules/` directory. You can apply individual detection rules directly to the cluster:

   ```bash
   kubectl apply -f config/rules/example-detection-rule.yaml -n <namespace>
   ```

- However, when running Helm, this process is automated, and all enabled rules in the `config/rules/` directory are deployed.

- The `config/exclusions/` directory contains exclusion lists to filter false positives. These exclusion lists can be configured in each rule using the `exclusionsPath` field. Example:

```yaml
name: example-rule
uid: e917a1fe-aa8e-424a-b81f-1f8e518c7282
status: development
confidence: low
enabled: true
schedule: "0 * * * *"
queryEngine: opensearch.rel
exclusionsPath: /app/exclusion/example-rule.yaml
publishers:
  - pubsub.alerts
language: SQL
...
```

---

### Final Notes

- **Automation**: While this guide includes manual steps for building and deploying Venator, it is recommended to automate the build and deployment process using a CI/CD tool like GitLab CI or GitHub Actions.
  
- **Customization**: Modify the configurations to suit your environment:
  - Helm configs: `config/values.yaml`, `config/templates/cronjob.yaml`
  - Venator configs: `config/files/global_config.yaml`, `config/rules/`, and `config/exclusions/`

- **Testing**: Always test your configuration with `--dry-run` before applying changes to your live environment.