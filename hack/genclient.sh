#!/usr/bin/env bash

go mod tidy
go mod vendor
retVal=$?
if [ $retVal -ne 0 ]; then
    exit $retVal
fi

set -e
TMP_DIR=$(mktemp -d)
mkdir -p "${TMP_DIR}"/src/github.com/ysicing/cloudflow
cp -r ./{apis,hack,vendor,go.mod} "${TMP_DIR}"/src/github.com/ysicing/cloudflow

(cd "${TMP_DIR}"/src/github.com/ysicing/cloudflow; \
    GOPATH=${TMP_DIR} GO111MODULE=off /bin/bash vendor/k8s.io/code-generator/generate-groups.sh all \
    github.com/ysicing/cloudflow/pkg/client github.com/ysicing/cloudflow/apis "apps:v1beta1 jobs:v1beta1" -h ./hack/boilerplate.go.txt ;
    )

rm -rf ./pkg/client
mkdir ./pkg/client
tree "${TMP_DIR}"/src/github.com/ysicing/cloudflow/pkg/client/
mv "${TMP_DIR}"/src/github.com/ysicing/cloudflow/pkg/client/* ./pkg/client/
rm -rf ${TMP_DIR}
rm -rf vendor
