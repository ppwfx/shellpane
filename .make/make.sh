#!/usr/bin/env bash

set -eox pipefail

function run-backend() {
  export SPECS_YAML="$(cat .make/specs.yaml)"

  go run cmd/main.go --http-addr 0.0.0.0:8888
}

function test() {
  go test ./...
}

function test-release() {
  git tag v0.0.0
  (goreleaser --skip-publish --rm-dist -f .make/.goreleaser.yml && git tag -d v0.0.0) || (git tag -d v0.0.0 && exit 1)
}

function run-frontend() {
  (cd internal/communication/web && yarn install && REACT_APP_SHELLPANE_HOST=http://localhost:8888 yarn start)
}

function build() {
  (cd internal/communication/web && yarn install && yarn build)
  go build -o shellpane cmd/main.go
}

function release() {
  goreleaser --rm-dist -f .make/.goreleaser.yml
}