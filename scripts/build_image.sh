#!/bin/bash
# Run this script from the directory containing Dockerfile.
gcloud artifacts repositories create venator-repo --repository-format=docker --location=us-central1 || true
gcloud builds submit --region=us-central1 --tag us-central1-docker.pkg.dev/example-project/venator-repo/venator-image:lastest
