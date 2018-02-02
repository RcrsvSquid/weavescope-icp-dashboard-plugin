#!/bin/bash
set -ex

TAG=${1:-latest}
export IMAGE_NAME="mycluster.icp:8500/default/weavescope-icp-dashboard-plugin:$TAG"

docker build -t $IMAGE_NAME .
docker push $IMAGE_NAME

kubectl delete -f deploy.yml
kubectl apply -f deploy.yml
