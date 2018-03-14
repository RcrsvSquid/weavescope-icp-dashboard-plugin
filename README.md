# ICP Weave Scope Plugin 
[![Build Status](https://travis.ibm.com/IBMPrivateCloud/weavescope-icp-plugin.svg?token=FQtRyxd2oucrshZSEEqZ&branch=master)](https://travis.ibm.com/IBMPrivateCloud/weavescope-icp-plugin)

A [weave scope plugin](https://www.weave.works/docs/scope/latest/plugins/) that computes links into the ICP dashboard for each kubernetes object deployed in the environment.

# Getting started
1. Configure kubectl
1. [Install Weave Scope](https://www.weave.works/docs/scope/latest/installing/#k8s)
1. Change the value of "DOCKER_USER" in **build.sh**
1. Run
```bash
$ ./build.sh
```
