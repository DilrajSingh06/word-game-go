param (
    [Parameter(Mandatory=$true)]
    [string]$DockerHubUser,

    [Parameter(Mandatory=$true)]
    [string]$DockerHubToken
)

$ErrorActionPreference = "Stop"

Write-Host "=== 1. Creating Namespaces (qa, stage, prod) ===" -ForegroundColor Cyan
kubectl apply -f namespaces.yaml

$namespaces = @("qa", "stage", "prod")

foreach ($ns in $namespaces) {
    Write-Host "=== Setting up namespace: $ns ===" -ForegroundColor Green

    # Recreate dockerhub secret
    kubectl delete secret dockerhub -n $ns --ignore-not-found
    kubectl create secret docker-registry dockerhub `
        --docker-username=$DockerHubUser `
        --docker-password=$DockerHubToken `
        --docker-email="unused@example.com" `
        -n $ns

    # Apply Postgres and Redis
    kubectl apply -f "$ns/postgres.yaml"
    kubectl apply -f "$ns/redis.yaml"

    # Apply App with DockerHub username replaced
    $appYaml = Get-Content "$ns/app.yaml" -Raw
    $appYaml = $appYaml -replace 'DOCKERHUB_USER', $DockerHubUser
    $appYaml | kubectl apply -f -
}

Write-Host "`n=== All namespaces configured and deployed! ===" -ForegroundColor Cyan
Write-Host "Verify pods status with: kubectl get pods -n qa; kubectl get pods -n stage; kubectl get pods -n prod" -ForegroundColor Yellow
