# ICP Dashboard Weave Scope Plugin

A [weave scope plugin](https://www.weave.works/docs/scope/latest/plugins/) that
links kubernetes objects to their dashboard view in the ICP dashboard

# Getting started
1. Configure kubectl
1. [Install Weave Scope](https://www.weave.works/docs/scope/latest/installing/#k8s)
1. set $DOCKER_USER in your environment
1. Change the image name deploy the configmap needed in `deploy.yml`
1. Run
```bash
$ ./build.sh
```

##  Faster Dev Builds
The Dockerfile will always `go get` the kubernetes dependency. This makes the
build take much longer than need be. For this reason I've included the
`.local_build/` folder with a separate Dockerfile. The script
`.local_build/build.sh` is the same as the top level version except it complies
the go binary outside of the container and therefore can reuse the dependencies.

To use
1. Run `go get k8s.io/client-go/...`
1. Run `.local_build/build.sh`

## Viewing the report
I've added a test case that logs the weave report so that you may verify it's
contents. This is a quick solution that I have found helpful in development.

Run `go test -v *.go`
