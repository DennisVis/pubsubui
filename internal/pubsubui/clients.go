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

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
)

func createClients(ctx context.Context, projectIDs []string) (map[string]*pubsub.Client, error) {
	clients := make(map[string]*pubsub.Client)

	for _, projectID := range projectIDs {
		logWithPrefix("clients: creating: for project %q", projectID)

		client, err := pubsub.NewClient(ctx, projectID)
		if err != nil {
			return nil, errors.Wrapf(err, "clients: could not create for project %q", projectID)
		}

		clients[projectID] = client

		logWithPrefix("clients: created for project %q", projectID)
	}

	return clients, nil
}
