#!/bin/bash
# If you don't use Helm to automate this, run this script to create ConfigMaps from the YAML configuration files in the `rules` directory.
for file in ../rules/*.yaml; do
    basename=$(basename $file .yaml)
    kubectl delete configmap $basename-config || true
    kubectl create configmap $basename-config --from-file=$file
done