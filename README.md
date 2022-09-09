# cloudflow

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/ysicing/cloudflow?filename=go.mod&style=flat-square)
![GitHub commit activity](https://img.shields.io/github/commit-activity/w/ysicing/cloudflow?style=flat-square)
![GitHub](https://img.shields.io/badge/license-AGPL-blue)
[![Releases](https://img.shields.io/github/release-pre/ysicing/cloudflow.svg)](https://github.com/ysicing/cloudflow/releases)
[![docs](https://img.shields.io/badge/docs-done-green)](https://blog.ysicing.net/)

## Description

自用CRD练习。

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster

1. Install Instances of Custom Resources:

```sh
kubectl apply -f hack/deploy/deploy.yaml
# example
kubectl apply -f config/samples/web
```

## SDK

```bash
go get -u github.com/ysicing/cloudflow@latest
```

## License

Copyright 2022 ysicing(i@ysicing.me).
