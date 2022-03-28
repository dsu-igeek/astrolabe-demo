#
# Copyright 2019 VMware, Inc..
# SPDX-License-Identifier: Apache-2.0
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

export GO111MODULE=on
export GOFLAGS=-mod=readonly

all: build

build:  cmds plugins

cmds: astrolabe astrolabe_server astrolabe_repo_server

astrolabe:
	cd cmd/astrolabe; go build

astrolabe_server:
	cd cmd/astrolabe_server; go build

astrolabe_repo_server:
	cd cmd/astrolabe_repo_server; go build

#
# TODO - move the minio main here, needs changes to use Astrolabe plugins to avoid circular dependency
#
astrolabe_minio:
	cd ../../minio/minio; make

plugins: psql_plugin

psql_plugin:
	cd plugins/psql; go build

astrolabe_minio_container: astrolabe_minio
	mkdir -p docker_build/astrolabe_minio
	cp -R docker/astrolabe_minio/* docker_build/astrolabe_minio
	mkdir -p docker_build/astrolabe_minio/overrides/bin
	cp ../../minio/minio/minio docker_build/astrolabe_minio/overrides/bin
	cd docker_build/astrolabe_minio; docker build -t astrolabe_minio_container .
