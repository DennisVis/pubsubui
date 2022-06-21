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
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/pkg/errors"
)

type PubSubMessage struct {
	ID          string            `json:"id"`
	Data        json.RawMessage   `json:"data"`
	PublishTime time.Time         `json:"publishTime"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

type SSEEvent struct {
	ID    string
	Event string
	Data  []byte
}

func sseEventFromPubSubMessage(msg *pubsub.Message) (*SSEEvent, error) {
	bts, err := json.Marshal(&PubSubMessage{
		ID:          msg.ID,
		Data:        msg.Data,
		PublishTime: msg.PublishTime,
		Attributes:  msg.Attributes,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal pubsub message to JSON")
	}

	return &SSEEvent{
		ID:    msg.ID,
		Event: "message",
		Data:  bts,
	}, nil
}

func (ssee SSEEvent) String() string {
	sb := strings.Builder{}

	if ssee.ID != "" {
		sb.WriteString("id: ")
		sb.WriteString(ssee.ID)
		sb.WriteString("\n")
	}
	if ssee.Event != "" {
		sb.WriteString("event: ")
		sb.WriteString(ssee.Event)
		sb.WriteString("\n")
	}
	if ssee.Data != nil {
		sb.WriteString("data: ")
		sb.WriteString(strings.ReplaceAll(string(ssee.Data), "\n", ""))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	return sb.String()
}

type SSEClient = chan SSEEvent

type ServerSSE struct {
	subscribeCh   chan SSEClient
	unSubscribeCh chan SSEClient
}

func (srv *ServerSSE) handle(ctx context.Context) {
	clients := make(map[SSEClient]bool)

	for {
		select {
		case sub := <-srv.subscribeCh:
			clients[sub] = true
		case unsub := <-srv.unSubscribeCh:
			delete(clients, unsub)
		case <-ctx.Done():
			for client := range clients {
				close(client)
				delete(clients, client)
			}
		}
	}
}

func (srv *ServerSSE) Subscribe(w http.ResponseWriter, r *http.Request, messageCh chan *pubsub.Message) {
	ctx := r.Context()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	client := make(SSEClient)
	srv.subscribeCh <- client
	defer func() {
		srv.unSubscribeCh <- client
	}()

	for {
		select {
		case msg := <-messageCh:
			event, err := sseEventFromPubSubMessage(msg)
			if err != nil {
				logWithPrefix("could not convert pubsub message to SSE event: %+v\n", err)
				continue
			}

			w.Write([]byte(event.String()))
			flusher.Flush()

			msg.Ack()
		case <-ctx.Done():
			return
		}
	}
}
