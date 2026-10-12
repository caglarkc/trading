#!/bin/sh
set -eu
cd /opt/trading/staging-src
docker run --rm --network trading-staging_database \
 --env-file /etc/trading/staging/test.env --env CGO_ENABLED=0 --env GOPROXY=off \
 --mount type=bind,src=/opt/trading/staging-src,dst=/src \
 --mount type=volume,src=trading-go-mod,dst=/go/pkg/mod \
 --mount type=volume,src=trading-go-build,dst=/root/.cache/go-build \
 -w /src golang@sha256:8a5910f31396cd4d89662f56c68b3ae31d374308270a1c3bd96672ee5ed43414 \
 go test -timeout 60s -v ./internal/store ./internal/jobs
