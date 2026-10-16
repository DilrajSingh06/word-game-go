#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -lt 2 ]; then
  echo "Usage: ./setup-k8s.sh <DOCKERHUB_USER> <DOCKERHUB_TOKEN>"
  exit 1
fi

DOCKERHUB_USER="$1"
DOCKERHUB_TOKEN="$2"

echo "=== 1. Creating Namespaces (qa, stage, prod) ==="
kubectl apply -f namespaces.yaml

for ns in qa stage prod; do
  echo "=== Setting up namespace: $ns ==="

  # Create / update Docker Hub pull secret
  kubectl delete secret dockerhub -n "$ns" --ignore-not-found
  kubectl create secret docker-registry dockerhub \
    --docker-username="$DOCKERHUB_USER" \
    --docker-password="$DOCKERHUB_TOKEN" \
    --docker-email="unused@example.com" \
    -n "$ns"

  # Deploy Postgres and Redis
  kubectl apply -f "$ns/postgres.yaml"
  kubectl apply -f "$ns/redis.yaml"

  # Deploy App with Docker Hub username substituted
  sed "s|DOCKERHUB_USER|$DOCKERHUB_USER|g" "$ns/app.yaml" | kubectl apply -f -
done

echo ""
echo "=== Deployment initiated for all namespaces! ==="
echo "Check pods with: kubectl get pods -A"
