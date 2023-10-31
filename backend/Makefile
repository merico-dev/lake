# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# NOTE: you have to use tabs in this file for make. Not spaces.
# https://stackoverflow.com/questions/920413/make-error-missing-separator
# https://tutorialedge.net/golang/makefiles-for-go-developers/

SHA = $(shell git show -s --format=%h)
TAG ?= $(shell git tag --points-at HEAD)
IMAGE_REPO ?= "apache"
VERSION = $(TAG)@$(SHA)
PYTHON_DIR ?= "./python"


all: build

go-dep:
	go install github.com/vektra/mockery/v2@v2.20.0
	go install github.com/swaggo/swag/cmd/swag@v1.8.12
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3
	go install github.com/atombender/go-jsonschema/cmd/gojsonschema@latest

python-dep:
	pip install -r python/requirements.txt

dep: go-dep python-dep

swag:
	scripts/swag.sh

build-plugin:
	scripts/build-plugins.sh

build-server: swag
	scripts/build-server.sh

build-python: #don't mix this with the other build commands
	scripts/build-python.sh

build: build-plugin build-server

run:
	go run server/main.go

dev: build-plugin build-python run

godev:
	DISABLED_REMOTE_PLUGINS=true make build-plugin run

debug: build-plugin-debug
	dlv debug server/main.go

mock:
	rm -rf mocks
	mockery --recursive --keeptree --dir=./core --output=./mocks/core --unroll-variadic=false --name='.*'
	mockery --recursive --keeptree --dir=./helpers --output=./mocks/helpers --unroll-variadic=false --name='.*'

test: unit-test e2e-test

unit-test: mock unit-test-go unit-test-python

unit-test-go:
	scripts/unit-test-go.sh

build-pydevlake:
	poetry install -C python/pydevlake

unit-test-python: build-pydevlake
	sh python/build.sh python/test &&\
	sh ./python/run_tests.sh

e2e-test-go-plugins:
	scripts/e2e-test-go-plugins.sh

e2e-test:
	scripts/e2e-test.sh

lint:
	golangci-lint run

fmt:
	find . -name \*.go | xargs gofmt -s -w -l

clean:
	@rm -rf bin

build-server-image:
	docker build -t $(IMAGE_REPO)/devlake:$(TAG) --build-arg TAG=$(TAG) --build-arg SHA=$(SHA) --file ./Dockerfile .

migration-script-lint:
	go run core/migration/linter/main.go -- $$(find . -path '**/migrationscripts/**.go')
