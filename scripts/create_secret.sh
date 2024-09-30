#!/bin/bash
# Run this script to create Secrets from the env file.
#
#  .vcfg.env    See dot-vcfg.env for example. Do not commit the env file to git!  
if [ -e .vcfg.env ]
then
  kubectl create secret generic venator-prod-secret --from-env-file=.vcfg.env -n venator
else
  echo Please create .vcfg.env before running this script. See dot-vcfg.env for example.
fi
