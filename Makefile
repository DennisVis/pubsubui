# Copyright 2022 Dennis Vis
# 
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

all: build

default: build

build_js:
	cd web/pubsubui && npm ci
	cd web/pubsubui && npm run build

build_go:
	GOOS=linux GOARCH=amd64 go build -o ./bin/pubsubui-linux cmd/pubsubui/main.go
	GOOS=darwin GOARCH=amd64 go build -o ./bin/pubsubui-darwin cmd/pubsubui/main.go
	GOOS=windows GOARCH=amd64 go build -o ./bin/pubsubui.exe cmd/pubsubui/main.go

build_docker:
	docker build -t dennisvis/pubsubui:latest -t dennisvis/pubsubui:$$(git tag -l --sort=-creatordate | head -n 1) .

build: build_js build_go
