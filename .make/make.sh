#!/usr/bin/env bash

set -eox pipefail

function run-backend() {
  export SPECS_YAML="$(cat .make/specs.yaml)"

  go run cmd/main.go --http-addr 0.0.0.0:8888
}

function test() {
  go test ./...
}

function run-frontend() {
  (cd internal/communication/web && yarn install && REACT_APP_SHELLPANE_HOST=http://localhost:8888 yarn start)
}

function build() {
  (cd internal/communication/web && yarn install && yarn build)
  go build -o shellpane cmd/main.go
}