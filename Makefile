SOURCE_MAKE=. ./.make/make.sh
SHELL := /bin/bash

run-backend:
	@${SOURCE_MAKE} && run-backend

run-frontend:
	@${SOURCE_MAKE} && run-frontend

build:
	@${SOURCE_MAKE} && build

test:
	@${SOURCE_MAKE} && test

test-release:
	@${SOURCE_MAKE} && test-release

release:
	@${SOURCE_MAKE} && release