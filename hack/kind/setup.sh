#!/bin/bash

docker pull ccr.ccs.tencentyun.com/k7scn/kindnode:v1.23.6

kind create cluster --image ccr.ccs.tencentyun.com/k7scn/kindnode:v1.23.6 --kubeconfig ./kind.kubeconfig --config ./hack/kind/kind.yaml
