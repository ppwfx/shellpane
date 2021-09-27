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