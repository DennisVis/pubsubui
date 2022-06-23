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
	"strings"
	"syscall"

	"cloud.google.com/go/pubsub"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const AppName = "pubsubui"

func doAppSetup(
	ctx context.Context,
	projectIDs []string,
	configFilePath string,
	projectsCh chan<- []string,
	clientsCh chan<- map[string]*pubsub.Client,
	topicsCh chan<- Topics,
	topicsCreatedCh chan<- struct{},
) error {
	logWithPrefix("setup: starting")

	skipTopicCreation := false

	var topics Topics
	if configFilePath == "" {
		skipTopicCreation = true
		logWithPrefix("setup: no config file path provided, skipping topic creation")
	} else {
		rdr, err := os.Open(configFilePath)
		if err != nil {
			return errors.Wrapf(err, "setup: could not open config file location %q", configFilePath)
		}

		parsedTopics, err := parseTopics(rdr)
		if err != nil {
			rdr.Close()
			return errors.Wrap(err, "setup: could not parse topics config")
		}
		rdr.Close()

		topics = parsedTopics
	}

	topicsCh <- topics

	allProjectIDs := deduplicateStrings(append(projectIDs, topics.ProjectIDs()...))
	if len(allProjectIDs) == 0 {
		return errors.New("setup: no GCP projects configured")
	}

	projectsCh <- allProjectIDs

	logWithPrefix("setup: supporting the following Google Cloud Platform projects: %s", strings.Join(projectIDs, ", "))

	clients, err := createClients(ctx, allProjectIDs)
	if err != nil {
		return errors.Wrap(err, "setup: could not create Pub/Sub clients")
	}

	clientsCh <- clients

	if !skipTopicCreation {
		err = createTopics(ctx, clients, topics)
		if err != nil {
			return errors.Wrap(err, "setup: could not create Pub/Sub topics")
		}
	}

	topicsCreatedCh <- struct{}{}

	logWithPrefix("setup: finished")

	return nil
}

func RunAppWithContext(ctx context.Context, additionalRouterConfigs ...func(chi.Router)) error {
	logWithPrefix("application: starting")

	cfg, err := newConfig()
	if err != nil {
		return errors.Wrap(err, "application: could not create config")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	projectsCh := make(chan []string)
	clientsCh := make(chan map[string]*pubsub.Client)
	topicsCh := make(chan Topics)
	topicsCreatedCh := make(chan struct{})

	srvr := newServer(ctx, projectsCh, clientsCh, topicsCh, topicsCreatedCh, additionalRouterConfigs...)

	setupGroup := errgroup.Group{}
	setupGroup.Go(func() error {
		defer close(projectsCh)
		defer close(clientsCh)
		defer close(topicsCh)
		defer close(topicsCreatedCh)

		err := doAppSetup(ctx, cfg.projectIDs, cfg.configFilePath, projectsCh, clientsCh, topicsCh, topicsCreatedCh)
		if err != nil {
			return errors.Wrap(err, "setup: failed")
		}

		return nil
	})

	runGroup := errgroup.Group{}
	runGroup.Go(func() error {
		err := srvr.Start(ctx, cfg.host, cfg.port)
		if err != nil {
			return errors.Wrap(err, "application: server: stopped with error")
		}

		return nil
	})

	err = setupGroup.Wait()
	if err != nil {
		return errors.Wrap(err, "application: failed to start")
	}

	err = runGroup.Wait()
	if err != nil {
		return errors.Wrap(err, "application: stopped with error")
	}

	logWithPrefix("application: stopped")

	return nil
}

func RunApp(additionalRouterConfigs ...func(chi.Router)) error {
	return RunAppWithContext(context.Background(), additionalRouterConfigs...)
}
