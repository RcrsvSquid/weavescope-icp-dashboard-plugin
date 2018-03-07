#!/bin/bash
set -ex

DOCKER_USER=ycao

TAG=${1:-latest}
export IMAGE_NAME="$DOCKER_USER/weavescope-icp-dashboard-plugin:$TAG"

docker build -t $IMAGE_NAME .
docker push $IMAGE_NAME

# kubectl delete -f deploy.yml
# kubectl apply -f deploy.yml
