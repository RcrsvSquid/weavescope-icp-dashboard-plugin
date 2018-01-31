#!/bin/bash
set -e # bail on error
set -x # print each command

TAG=${1:-latest}
IMAGE_NAME="mycluster.icp:8500/default/iowait-modification:$TAG"

docker build -t $IMAGE_NAME .
docker push $IMAGE_NAME

# kubectl delete -f deploy.yml
kubectl apply -f deploy.yml
# kubectl set image ds/weavescope-testiowait-plugin weavescope-testiowait-plugin=$IMAGE_NAME
