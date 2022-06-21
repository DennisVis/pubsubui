// Copyright 2022 Dennis Vis
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pubsubui

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"cloud.google.com/go/pubsub"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	AppName   = "pubsubui"
	LogPrefix = AppName + ": "
)

const (
	envKeyHost     = "PUBSUBUI_HOST"
	envKeyPort     = "PUBSUBUI_PORT"
	envKeyProjects = "GOOGLE_CLOUD_PROJECTS"
	envKeyConfig   = "PUBSUBUI_CONFIG"
)

func doAppSetup(
	ctx context.Context,
	projectIDs []string,
	configFilePath string,
	projectsCh chan<- []string,
	clientsCh chan<- map[string]*pubsub.Client,
	topicsCh chan<- Topics,
	topicsCreatedCh chan<- struct{},
) error {
	rdr, err := os.Open(configFilePath)
	if err != nil {
		return errors.Wrapf(err, "setup: could not open config file location %q", configFilePath)
	}

	topics, err := parseTopics(rdr)
	if err != nil {
		rdr.Close()
		return errors.Wrap(err, "setup: could not parse topics config")
	}
	rdr.Close()

	topicsCh <- topics

	projectIDs = deduplicateStrings(append(projectIDs, topics.ProjectIDs()...))

	projectsCh <- projectIDs

	logWithPrefix("setup: supporting the following Google Cloud Platform projects: %s", strings.Join(projectIDs, ", "))

	clients, err := createClients(ctx, projectIDs)
	if err != nil {
		return errors.Wrap(err, "setup: could not create Pub/Sub clients")
	}

	clientsCh <- clients

	err = createTopics(ctx, clients, topics)
	if err != nil {
		return errors.Wrap(err, "setup: could not create Pub/Sub topics")
	}

	topicsCreatedCh <- struct{}{}

	return nil
}

func RunAppWithContext(ctx context.Context, additionalRouterConfigs ...func(chi.Router)) error {
	logWithPrefix("application: starting")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := envOrDefault(envKeyHost, "0.0.0.0")
	portStr := envOrDefault(envKeyPort, "8080")
	configFile := envOrDefault(envKeyConfig, "config.yaml")

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return errors.Wrapf(err, "application: invalid port number: %s", portStr)
	}

	projectIDsStr := envOrDefault(envKeyProjects, "")
	projectIDs := filterEmptyStrings(strings.Split(projectIDsStr, ","))

	g := errgroup.Group{}

	projectsCh := make(chan []string)
	clientsCh := make(chan map[string]*pubsub.Client)
	topicsCh := make(chan Topics)
	topicsCreatedCh := make(chan struct{})

	srvr := newServer(ctx, projectsCh, clientsCh, topicsCh, topicsCreatedCh, additionalRouterConfigs...)

	g.Go(func() error {
		defer close(projectsCh)
		defer close(clientsCh)
		defer close(topicsCh)
		defer close(topicsCreatedCh)

		err := doAppSetup(ctx, projectIDs, configFile, projectsCh, clientsCh, topicsCh, topicsCreatedCh)
		if err != nil {
			return errors.Wrap(err, "application: setup failed")
		}

		return nil
	})

	g.Go(func() error {
		err := srvr.Start(ctx, host, uint(port))
		if err != nil {
			return errors.Wrap(err, "application: server: stopped with error")
		}

		return nil
	})

	err = g.Wait()
	if err != nil {
		return errors.Wrap(err, "application: stopped with error")
	}

	logWithPrefix("application: stopped")

	return nil
}

func RunApp(additionalRouterConfigs ...func(chi.Router)) error {
	return RunAppWithContext(context.Background(), additionalRouterConfigs...)
}
