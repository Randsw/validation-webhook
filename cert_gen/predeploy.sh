#!/bin/sh

# Exit on error.
set -e

usage() {
    echo 'Usage: predeploy.sh <serviceName> <namespace> <secretName>'
}

if [ "$#" -ne 3 ]; then
    usage
    exit 1
fi

service=$1
namespace=$2
secret=$3

# Create namespace if not exists.
kubectl create namespace $namespace || true

# Populate secrets from certificate file and key.
cert_gen/generate-secret.sh $service $namespace $secret
