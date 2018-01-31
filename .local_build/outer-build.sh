#!/bin/bash
set -ex

TAG=${1:-latest}
IMAGE_NAME="mycluster.icp:8500/default/iowait-modification:$TAG"

CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

docker build --rm -t $IMAGE_NAME -f .local_build/OuterBuild .
# docker run $IMAGE_NAME
docker push $IMAGE_NAME

kubectl delete -f deploy.yml
kubectl apply -f deploy.yml
