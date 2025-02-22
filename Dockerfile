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

FROM node:18-alpine3.15 AS js-build
WORKDIR /build
COPY web/pubsubui/ ./
RUN npm ci
RUN npm run build

FROM golang:1.18.3-alpine3.16 AS go-build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY ./cmd/ ./cmd/
COPY ./internal/ ./internal/
COPY --from=js-build /build/dist/ ./web/pubsubui/dist
COPY ./web/pubsubui/pubsubui.go ./web/pubsubui/pubsubui.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/pubsubui ./cmd/pubsubui/main.go

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=go-build /build/bin/pubsubui ./pubsubui
ENTRYPOINT ["./pubsubui"]
