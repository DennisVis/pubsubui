# Pub/Sub UI

[![Go Report Card](https://goreportcard.com/badge/github.com/DennisVis/pubsubui)](https://goreportcard.com/report/github.com/DennisVis/pubsubui)

Pub/Sub UI is graphical user interface for managing Google Cloud Pub/Sub.

## Overview
This application provides the following features:

- Switching between multiple GCP projects
- Browsing all Pub/Sub topics created within a GCP project
- Subscribing to a topic and receiving messages as they come in
- Publishing messages to a topic
- Using pre-defined message payload for publishing
- Creating new topics
- Creating new subscriptions

## Configuration
The following environment variables are supported:

| Variable                         | Usage                                                          | Default value |
|----------------------------------|----------------------------------------------------------------|---------------|
| `PUBSUBUI_HOST`                  | Listening HTTP host                                            | `0.0.0.0`     |
| `PUBSUBUI_PORT`                  | Listening HTTP port                                            | `8080`        |
| `PUBSUBUI_CONFIG`                | Config file path (see below)                                   | `config.yaml` |
| `GOOGLE_CLOUD_PROJECTS`          | Comma-separated list of GCP project IDs (mind the plural form) | _none_        |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to Google Cloud Platform JSON credentials file            | _none_        |
| `PUBSUB_EMULATOR_HOST`           | Address of the Pub/Sub emulator (see below)                    | _none_        |

The `GOOGLE_CLOUD_PROJECTS` is only needed if the `config.yaml` does not contain all the target GCP project IDs.

The `PUBSUB_EMULATOR_HOST` is optional and functions the same as described here: 
https://cloud.google.com/pubsub/docs/emulator#manually_setting_the_variables

If `PUBSUB_EMULATOR_HOST` is not set the application will attempt to connect to the actual GCP projects. In this case 
the `GOOGLE_APPLICATION_CREDENTIALS` will have to be set, otherwise authentication will fail.

### The `config.yaml` file
The application can be configured to automatically create topics and their subscriptions, as well as pre-defined 
message payloads to send to these topics. This is done within a YAML file, an example can be found below.

```yaml
topics:
- name: my-topic
  project: my-gcp-project
  subscriptions:
  - my-first-subscription
  - my-second-subscription
- name: my-other-topic
  project: other-gcp-project
  subscriptions:
  - my-other-subscription
  payloads:
  - name: hello
    payload: |
      {
        "hello": "world"
      }
  - name: hello-again
    payload: |
      {
        "hello": "world again"
      }
- name: my-last-topic
```

- All topics specified will be automatically created
- All subscriptions will be automatically created on the topic they are defined under
- Configured payloads will be presented in the UI for the topic they are defined under
- All project IDs will be extracted and be made selectable within the UI

## Usage

### Using a binary
Download latest release for your OS from the [releases](https://github.com/DennisVis/pubsubui/releases) page and make it available on your `$PATH`.

Then:

```bash
export PUBSUBUI_CONFIG="/path/to/config.yaml"

# To use local emulator
export PUBSUB_EMULATOR_HOST=localhost:8085

# To use actual Pub/Sub service
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/application_default_credentials.json"

# When needing to manage GCP projects which are not in the config.yaml
export GOOGLE_CLOUD_PROJECTS="my-first-gcp-project,my-second-gcp-project"

pubsubui
```

Then open http://localhost:8080.

### Using a Docker image
When running with the emulator:

```bash
docker pull dennisvis/pubsubui

docker run \
  --name="pubsubui" \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/config/config.yaml \
  -e PUBSUBUI_CONFIG=/config/config.yaml \
  -e PUBSUB_EMULATOR_HOST=host.docker.internal:8085 \
  -e GOOGLE_CLOUD_PROJECTS="my-first-gcp-project,my-second-gcp-project" \
  dennisvis/pubsubui
```

When running using the actual cloud service:

```bash
docker pull dennisvis/pubsubui

docker run \
  --name="pubsubui" \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/config/config.yaml \
  -e PUBSUBUI_CONFIG=/config/config.yaml \
  -e GOOGLE_CLOUD_PROJECTS="my-first-gcp-project,my-second-gcp-project" \
  -v /path/to/application_default_credentials.json:/config/application_default_credentials.json \
  -e GOOGLE_APPLICATION_CREDENTIALS="/config/application_default_credentials.json" \
  dennisvis/pubsubui
```

Then open http://localhost:8080.

### Running on Kubernetes
The application exposes both a `/healthy` and a `/ready` endpoint which should be used for a liveness and readiness 
probe respectively in your Kubernetes manifest.

```yaml
containers:
- name: pubsubui
  ...
  livenessProbe:
    httpGet:
      port: 8080
      path: /healthy
  readinessProbe:
    httpGet:
      port: 8080
      path: /ready
```

## Building

### Binaries

```bash
make
```

Linux, MacOS (Darwin) and Windows binaries will be available under the `bin/` directory.

### Docker image

```bash
make build_docker
```

## Credits
This project was forked from- and based on 
[ClickAndMortar/GoPubSub](https://github.com/ClickAndMortar/GoPubSub).
