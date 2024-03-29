# AGPL License
# Copyright 2022 ysicing(i@ysicing.me).

# Build the manager binary
FROM ysicing/god as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod

COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go

COPY apis/ apis/

COPY pkg/ pkg/

COPY Makefile Makefile

# Build
RUN make build-only

FROM ysicing/debian

WORKDIR /

COPY --from=builder /workspace/bin/cloudflow .

USER 65532:65532

ENTRYPOINT ["/cloudflow"]
