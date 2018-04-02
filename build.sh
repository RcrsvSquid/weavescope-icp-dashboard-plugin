#!/bin/bash
set -ex

# expect $DOCKER_USER
if [ -z ${DOCKER_USER+x} ];
    then echo 'ERROR: $DOCKER_USER is unset'; exit 1;
    else echo "\$DOCKER_USER='$DOCKER_USER'";
fi

TAG=${1:-latest}
IMAGE_NAME="$DOCKER_USER/weavescope-icp-dashboard-plugin:$TAG"

docker build --rm -t $IMAGE_NAME .
docker push $IMAGE_NAME

# kubectl delete -f deploy.yml
# kubectl apply -f deploy.yml
