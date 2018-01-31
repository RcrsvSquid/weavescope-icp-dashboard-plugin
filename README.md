# ICP Weave Scope Plugin

# Getting started

1. Create and push the docker image with:
```bash
$ docker build -t <IMAGE_NAME> .
$ docker push <IMAGE_NAME>
```

2. Edit the deploy.yml to point to your image
3. Deploy the application as a DaemonSet with:
```
$ kubectl apply -f deploy.yml
```
