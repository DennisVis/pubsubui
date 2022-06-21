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
	"io"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/yaml.v3"
)

var (
	timeoutTopicCreation        = time.Second * 15
	timeoutSubscriptionCreation = time.Second * 15
)

type MessagePayload struct {
	Name    string `yaml:"name"    json:"name"`
	Payload string `yaml:"payload" json:"payload"`
}

type Topic struct {
	ID            string           `yaml:"-"             json:"id"`
	Name          string           `yaml:"name"          json:"name"`
	ProjectID     string           `yaml:"project"       json:"projectId"`
	Subscriptions []string         `yaml:"subscriptions" json:"-"`
	Payloads      []MessagePayload `yaml:"payloads"      json:"payloads"`
}

func (t Topic) Key() string {
	return t.ProjectID + "/" + t.Name
}

type Topics struct {
	Topics []Topic `yaml:"topics" json:"topics"`
}

func (ts Topics) ProjectIDs() []string {
	projectIDs := make([]string, len(ts.Topics))

	for i, topic := range ts.Topics {
		projectIDs[i] = topic.ProjectID
	}

	return deduplicateStrings(projectIDs)
}

func (ts Topics) Payloads() map[string][]MessagePayload {
	payloads := make(map[string][]MessagePayload)

	for _, topic := range ts.Topics {
		payloads[topic.Key()] = append(payloads[topic.Key()], topic.Payloads...)
	}

	return payloads
}

func parseTopics(yamlFile io.Reader) (Topics, error) {
	var topics Topics
	err := yaml.NewDecoder(yamlFile).Decode(&topics)
	if err != nil {
		return Topics{}, errors.Wrap(err, "could not parse topics")
	}

	return topics, nil
}

func createSubscription(
	ctx context.Context,
	client *pubsub.Client,
	projectID string,
	topicName string,
	subscriptionName string,
) error {
	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeoutSubscriptionCreation))
	defer func() {
		logWithPrefix(
			"subscription: creating: timeout on %q for topic %q in project %q",
			subscriptionName,
			topicName,
			projectID,
		)
		cancel()
	}()

	logWithPrefix("subscription: creating: %q for topic %q in project %q", subscriptionName, topicName, projectID)

	topic := client.Topic(topicName)

	_, err := client.CreateSubscription(ctx, subscriptionName, pubsub.SubscriptionConfig{
		Topic: topic,
	})
	if status.Code(err) == codes.AlreadyExists {
		logWithPrefix(
			"subscription: already exists: %q for topic %q in project %q",
			subscriptionName,
			topicName,
			projectID,
		)
		return nil
	}
	if err != nil {
		return errors.Wrapf(
			err,
			"subscription: could not create: %q for topic %q in project %q",
			subscriptionName,
			topicName,
			projectID,
		)
	}

	logWithPrefix("subscription: created: %q for topic %q in project %q", subscriptionName, topicName, projectID)

	return nil
}

func createTopic(ctx context.Context, client *pubsub.Client, topicCfg Topic) error {
	dlctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeoutTopicCreation))
	defer func() {
		logWithPrefix("topic: creating: timeout on %q in project %q", topicCfg.Name, topicCfg.ProjectID)
		cancel()
	}()

	logWithPrefix("topic: creating: %q in project %q", topicCfg.Name, topicCfg.ProjectID)

	_, err := client.CreateTopic(dlctx, topicCfg.Name)
	if status.Code(err) == codes.AlreadyExists {
		logWithPrefix("topic: already exists: %q in project %q", topicCfg.Name, topicCfg.ProjectID)
		goto CreateSubscriptions
	}
	if err != nil {
		return errors.Wrapf(err, "topics: could not create %q in project %q", topicCfg.Name, topicCfg.ProjectID)
	}

	logWithPrefix("topic: created: %q in project %q", topicCfg.Name, topicCfg.ProjectID)

CreateSubscriptions:
	sg := errgroup.Group{}
	for _, sn := range topicCfg.Subscriptions {
		subName := sn

		sg.Go(func() error {
			return createSubscription(ctx, client, topicCfg.ProjectID, topicCfg.Name, subName)
		})
	}
	err = sg.Wait()
	if err != nil {
		return errors.Wrapf(
			err,
			"topic: could not create for topic %q in project %q",
			topicCfg.Name,
			topicCfg.ProjectID,
		)
	}

	return nil
}

func createTopics(ctx context.Context, clients map[string]*pubsub.Client, topics Topics) error {
	tg := errgroup.Group{}

	if len(topics.Topics) == 0 {
		logWithPrefix("topics: not configured in the config file, skipping creation", len(topics.Topics))
		return nil
	}

	logWithPrefix("topics: creating %d topics from config file", len(topics.Topics))

	for _, tcfg := range topics.Topics {
		topicCfg := tcfg

		client, ok := clients[topicCfg.ProjectID]
		if !ok {
			return errors.Errorf("no client configured for project %q", topicCfg.ProjectID)
		}

		tg.Go(func() error {
			return createTopic(ctx, client, topicCfg)
		})
	}
	err := tg.Wait()
	if err != nil {
		return errors.Wrap(err, "topics: could not create")
	}

	logWithPrefix("topics: all %d topics created", len(topics.Topics))

	return nil
}
