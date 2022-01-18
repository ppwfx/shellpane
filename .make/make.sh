#!/usr/bin/env bash

set -eox pipefail

function run-backend() {
#  export SHELLPANE_YAML="$(cat labelsapi.yaml)"

  go run cmd/main.go serve --shellpane-yaml-path .make/shellpane.yaml --http-addr 0.0.0.0:8080 --cors-origin http://localhost:3000
}

function test() {
  go test ./...
}

function run-frontend() {
  (cd internal/communication/web && yarn install && REACT_APP_SHELLPANE_HOST=http://localhost:8080 REACT_APP_CATEGORIES_CSS_HOST=http://localhost:8080 yarn start)
}

function build() {
  (cd internal/communication/web && yarn install && REACT_APP_CATEGORIES_CSS_HOST="" yarn build)
  go build -o shellpane cmd/main.go
}