## Kubernetes/Helm Deployment Guide

This guide outlines the steps to deploy Venator on a Kubernetes cluster using Helm. Follow these instructions to set up the environment, build the Docker image, configure Helm, and deploy the application.

### Prerequisites

- **Kubernetes Cluster**  
   Ensure you have a Kubernetes cluster set up and configured. You can use any preferred method (e.g., Terraform, GCP Console, etc.) to create the cluster.

- **Cluster Access**  
   Verify that you have access to your Kubernetes cluster and can authenticate to it. Ensure your `kubectl` context is configured correctly for the target cluster. You can check or switch contexts as needed with:

   ```bash
   kubectl config get-contexts  # List available contexts
   kubectl config current-context  # Check the current context
   kubectl config use-context <context-name>  # Switch to the appropriate context
   ```


### Build the Docker Image

Build the Venator Docker image by running the provided script. Ensure you're in the root directory containing the Dockerfile, and that you're using an Artifact Registry repository named `venator-repo` to store the built artifacts:

   ```bash
   ./scripts/build_image.sh
   ```

> **Note:** This is a manual build process. Automating it via a CI/CD pipeline (like GitLab CI or GitHub Actions) is recommended.

### Create Secrets

If you need to store credentials for your connectors, create Kubernetes secrets using the `scripts/create_secret.sh` script or another secret manager. These secrets are referenced in `config/files/global_config.yaml` and `config/templates/cronjob.yaml`. Ensure they are properly configured.

For example, to set the OpenSearch password locally:

```bash
read -s -r OPENSEARCH_PASSWORD
export OPENSEARCH_PASSWORD
```

In `global_config.yaml`, reference the environment variable:

```yaml
opensearch:
  instances:
    prod:
      url: https://opensearch:9200
      username: user
      password: ${OPENSEARCH_PASSWORD}
      insecureSkipVerify: true
```

For Kubernetes deployments, ensure the variable is properly referenced in `config/templates/cronjob.yaml`:

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
   The Helm chart configuration is located in the `config` directory. Before deploying, update the `values.yaml` file with your Docker image:

   ```yaml
   container:
     image: "your-docker-repo/venator-image:latest"  # Replace with your Docker image
   ```

2. **Test Helm Chart Locally (Dry Run)**  
   Test the Helm chart to ensure everything is set up correctly:

   ```bash
   cd config
   helm template . --dry-run  # Test chart rendering locally
   ```

   Use the `--debug` flag for more detailed output:

   ```bash
   helm install venator-test . --dry-run --debug  # Render chart locally without installing, checks for resource conflicts
   ```

3. **Install the Helm Chart**  
   Once everything is configured correctly, install the chart in the `venator` namespace:

   ```bash
   helm install venator-test . -n venator
   ```

4. **Upgrading the Chart**  
   For updates, use the Helm upgrade command:

   ```bash
   helm upgrade venator-test . -n venator
   ```

### Detection Rules

- Detection rules are defined as YAML files in the `config/rules/` directory. Using Helm, all enabled rules in this directory (along with exclusion lists and global configurations) are created as ConfigMaps in Kubernetes, and rules are automatically deployed as CronJobs.

Example snippet from `cronjob.yaml`:
```yaml
...
spec:
  timeZone: "Etc/UTC"
  schedule: {{ $cfg.schedule | quote }}
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: {{ $.Values.container.name }}
            imagePullPolicy: Always
            image: {{ $.Values.container.image }}
            volumeMounts:
            - name: rule-volume
              mountPath: /app/rule
            - name: config-volume
              mountPath: /app/config
            - name: exclusion-volume
              mountPath: /app/exclusion
            args:
              - "--rule-config"
              - "/app/rule/{{ $cfg.name }}.yaml"
              - "--global-config"
              - "/app/config/global_config.yaml"
...
```

To manually apply individual detection rules to the cluster without using Helm:

   ```bash
   kubectl apply -f config/templates/cronjob.yaml -n <namespace>
   ```

Alternatively, run Venator locally to execute individual rules without deploying to Kubernetes:

   ```bash
   ./venator --global-config config/files/global_config.yaml --rule-config config/rules/macos/macos-osascript-execution.yaml
   ```

- Exclusion lists, located in the `config/exclusions/` directory, help filter false positives. You can reference these exclusion lists in each rule using the `exclusionsPath` field. Example:

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

- **Automation**: Automating the build and deployment process via CI/CD tools like GitLab CI or GitHub Actions is highly recommended. An example [gitlab-ci.yml](../.gitlab-ci.yml) config is provided in the repo.
  
- **Customization**: Modify the configurations to suit your environment:
  - Helm: `config/values.yaml`, `config/templates/cronjob.yaml`
  - Venator: `config/files/global_config.yaml`, `config/rules/`, and `config/exclusions/`

- **Testing**: Always test your configurations with `--dry-run` before deploying to your live environment.