#!/bin/bash
set -ex

DOCKER_USER=sidneywibm

TAG=${1:-latest}
export IMAGE_NAME="$DOCKER_USER/weavescope-icp-dashboard-plugin:$TAG"

CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

docker build --rm -t $IMAGE_NAME -f .local_build/OuterBuild .
docker push $IMAGE_NAME

kubectl delete -f deploy.yml
kubectl apply -f deploy.yml
