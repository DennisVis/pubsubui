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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/go-chi/chi/v5"
	"github.com/lithammer/shortuuid/v4"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	statusHealthy = "Healthy"
	statusReady   = "Ready"
)

const (
	pageDefault           = "1"
	pageSizeStrDefault    = "10"
	queryParamKeyPage     = "page"
	queryParamKeyPageSize = "pageSize"
)

type Server struct {
	additionalRouterConfigs []func(chi.Router)
	statusMu                sync.Mutex
	projectIDs              []string
	projectsSet             bool
	clients                 map[string]*pubsub.Client
	clientsSet              bool
	payloads                map[string][]MessagePayload
	topicsSet               bool
	topicsCreated           bool
	topicsCache             map[string][]Topic
	sse                     *ServerSSE
}

func handleServerSetup(
	srv *Server,
	projectsCh <-chan []string,
	clientsCh <-chan map[string]*pubsub.Client,
	topicsCh <-chan Topics,
	topicsCreatedCh <-chan struct{},
) {
	isReady := func() bool {
		return srv.projectsSet && srv.clientsSet && srv.topicsSet && srv.topicsCreated
	}

Setup:
	for {
		select {
		case projects := <-projectsCh:
			srv.statusMu.Lock()

			srv.projectIDs = projects
			srv.projectsSet = true

			ready := isReady()

			srv.statusMu.Unlock()

			logWithPrefix("server: received GCP projects configuration")

			if ready {
				break Setup
			}
		case clients := <-clientsCh:
			srv.statusMu.Lock()

			srv.clients = clients
			srv.clientsSet = true

			ready := isReady()

			srv.statusMu.Unlock()

			logWithPrefix("server: received Google Cloud Pub/Sub clients")

			if ready {
				break Setup
			}
		case topics := <-topicsCh:
			srv.statusMu.Lock()

			srv.payloads = topics.Payloads()
			srv.topicsSet = true

			ready := isReady()

			srv.statusMu.Unlock()

			logWithPrefix("server: received topics configuration")

			if ready {
				break Setup
			}
		case <-topicsCreatedCh:
			srv.statusMu.Lock()

			srv.topicsCreated = true

			ready := isReady()

			srv.statusMu.Unlock()

			logWithPrefix("server: received topics created notification")

			if ready {
				break Setup
			}
		}
	}

	logWithPrefix("server: fully configured")
}

func newServer(
	ctx context.Context,
	projectsCh <-chan []string,
	clientsCh <-chan map[string]*pubsub.Client,
	topicsCh <-chan Topics,
	topicsCreatedCh <-chan struct{},
	additionalRouterConfigs ...func(chi.Router),
) *Server {
	srv := &Server{
		additionalRouterConfigs: additionalRouterConfigs,
		topicsCache:             make(map[string][]Topic),
	}

	go handleServerSetup(srv, projectsCh, clientsCh, topicsCh, topicsCreatedCh)

	srv.sse = &ServerSSE{
		subscribeCh:   make(chan SSEClient),
		unSubscribeCh: make(chan SSEClient),
	}

	go srv.sse.handle(ctx)

	return srv
}

type listProjectsResponse struct {
	Projects []string `json:"projects"`
}

type createTopicRequest struct {
	Name string `json:"name"`
}

type createTopicResponse struct {
	Topic Topic `json:"topic"`
}

type listTopicsResponse struct {
	ProjectID  string  `json:"projectId"`
	Topics     []Topic `json:"topics"`
	TotalItems uint    `json:"totalItems"`
	Page       uint    `json:"page"`
	PageSize   uint    `json:"pageSize"`
	TotalPages uint    `json:"totalPages"`
}

type publishMessageResponse struct {
	ProjectID string `json:"projectId"`
	MessageID string `json:"messageId"`
}

type createSubscriptionRequest struct {
	Name string `json:"name"`
}

type createSubscriptionResponse struct {
	Subscription struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		ProjectID string `json:"projectId"`
		TopicID   string `json:"topicId"`
	} `json:"subscription"`
}

func getQueryParamOrDefault(qry url.Values, key, def string) string {
	if !qry.Has(key) {
		return def
	}

	return qry.Get(key)
}

func topicNameFromTopicID(topicID string) string {
	split := strings.Split(topicID, "/")
	return split[len(split)-1]
}

func (srv *Server) status() string {
	srv.statusMu.Lock()
	defer srv.statusMu.Unlock()

	waitingFor := make([]string, 0)

	if !srv.projectsSet {
		waitingFor = append(waitingFor, "GCP projects configuration")
	}
	if !srv.clientsSet {
		waitingFor = append(waitingFor, "Pub/Sub client initialization")
	}
	if !srv.topicsSet {
		waitingFor = append(waitingFor, "topic configuration")
	}
	if !srv.topicsCreated {
		waitingFor = append(waitingFor, "topic creation")
	}

	if len(waitingFor) == 0 {
		return statusReady
	}

	return fmt.Sprintf("Waiting for %s", strings.Join(waitingFor, ", "))
}

func (srv *Server) Healthy(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(statusHealthy))
}

func (srv *Server) Ready(w http.ResponseWriter, r *http.Request) {
	status := srv.status()

	if status == statusReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Write([]byte(status))
}

func gRPCErrorCodeToHTTPStatus(c codes.Code) int {
	switch c {
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusRequestTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusAlreadyReported
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.ResourceExhausted:
		return http.StatusServiceUnavailable
	case codes.FailedPrecondition:
		return http.StatusPreconditionFailed
	case codes.Aborted:
		return http.StatusNotAcceptable
	case codes.OutOfRange:
		return http.StatusRequestedRangeNotSatisfiable
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DataLoss:
		return http.StatusInternalServerError
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		logWithPrefix("server: unknown gRPC error code: %s", c)
		return http.StatusInternalServerError
	}
}

func handleGoogleError(w http.ResponseWriter, actionTried string, err error) {
	grpcStatus := status.Convert(err)
	httpStatus := gRPCErrorCodeToHTTPStatus(grpcStatus.Code())

	if httpStatus >= 500 {
		logWithPrefix("server: could not %s: %+v", actionTried, err)
	}

	http.Error(w, grpcStatus.Message(), httpStatus)
}

func (srv *Server) ListProjects(w http.ResponseWriter, r *http.Request) {
	bts, err := json.Marshal(listProjectsResponse{
		Projects: srv.projectIDs,
	})
	if err != nil {
		logWithPrefix("server: %+v", errors.Wrap(err, "could not encode projects as JSON"))
		http.Error(w, "could not encode projects as JSON", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, "projects.json", time.Time{}, bytes.NewReader(bts))
}

func (srv *Server) CreateTopic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectID := chi.URLParam(r, "projectID")

	client, ok := srv.clients[projectID]
	if !ok {
		logWithPrefix("server: %+v", errors.Errorf("no client configured for project %q", projectID))
		http.Error(w, fmt.Sprintf("project %q not supported", projectID), http.StatusBadRequest)
		return
	}

	var req createTopicRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "could not decode create topic request", http.StatusBadRequest)
		return
	}

	topic, err := client.CreateTopic(ctx, req.Name)
	if err != nil {
		handleGoogleError(w, "create topic", err)
		return
	}

	topicID := topic.ID()
	topicName := topicNameFromTopicID(topicID)
	topicKey := fmt.Sprintf("%s/%s", projectID, topicName)
	payloads := srv.payloads[topicKey]

	newTopic := Topic{
		ID:        topicID,
		Name:      topicName,
		ProjectID: projectID,
		Payloads:  payloads,
	}

	srv.topicsCache[projectID] = append(srv.topicsCache[projectID], newTopic)

	bts, err := json.Marshal(createTopicResponse{newTopic})
	if err != nil {
		logWithPrefix("server: %+v", errors.Wrap(err, "could not encode create topic response as JSON"))
		http.Error(w, "could not encode create topic response as JSON", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, "topic.json", time.Time{}, bytes.NewReader(bts))
}

func (srv *Server) ListTopics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectID := chi.URLParam(r, "projectID")

	qry := r.URL.Query()

	pageStr := getQueryParamOrDefault(qry, queryParamKeyPage, pageDefault)
	page, err := strconv.ParseUint(pageStr, 10, strconv.IntSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid page %q", pageStr), http.StatusBadRequest)
		return
	}

	pageSizeStr := getQueryParamOrDefault(qry, queryParamKeyPageSize, pageSizeStrDefault)
	pageSize, err := strconv.ParseUint(pageSizeStr, 10, strconv.IntSize)
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid page size %q", pageSizeStr), http.StatusBadRequest)
		return
	}

	client, ok := srv.clients[projectID]
	if !ok {
		logWithPrefix("server: %+v", errors.Errorf("no client configured for project %q", projectID))
		http.Error(w, fmt.Sprintf("project %q not supported", projectID), http.StatusBadRequest)
		return
	}

	topics, ok := srv.topicsCache[projectID]
	if !ok {
		topicIt := client.Topics(ctx)
		for {
			topic, err := topicIt.Next()
			if errors.Is(err, iterator.Done) {
				break
			}
			if err != nil {
				handleGoogleError(w, " list topics", err)
				return
			}

			topicID := topic.ID()
			topicName := topicNameFromTopicID(topicID)
			topicKey := fmt.Sprintf("%s/%s", projectID, topicName)
			payloads := srv.payloads[topicKey]

			topics = append(topics, Topic{
				ID:        topicID,
				Name:      topicName,
				ProjectID: projectID,
				Payloads:  payloads,
			})
		}

		srv.topicsCache[projectID] = topics
	}

	totalItems := uint(len(topics))
	totalPages := uint(math.Ceil(float64(totalItems) / float64(pageSize)))
	offset := uint((page - 1) * pageSize)
	limit := offset + uint(pageSize)
	if limit > totalItems {
		limit = totalItems
	}

	bts, err := json.Marshal(listTopicsResponse{
		ProjectID:  projectID,
		Topics:     topics[offset:limit],
		TotalItems: totalItems,
		Page:       uint(page),
		PageSize:   uint(pageSize),
		TotalPages: totalPages,
	})
	if err != nil {
		logWithPrefix("server: %+v", errors.Wrap(err, "could not encode topics as JSON"))
		http.Error(w, "could not encode topics as JSON", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, "topics.json", time.Time{}, bytes.NewReader(bts))
}

func (srv *Server) Publish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectID := chi.URLParam(r, "projectID")
	topicID := chi.URLParam(r, "topicID")

	client, ok := srv.clients[projectID]
	if !ok {
		logWithPrefix("server: %+v", errors.Errorf("no client configured for project %q", projectID))
		http.Error(w, "project %q not supported", http.StatusBadRequest)
		return
	}

	msg, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logWithPrefix("server: %+v", errors.Wrap(err, "could not read message body"))
		http.Error(w, "could not read message body", http.StatusInternalServerError)
		return
	}

	topic := client.Topic(topicID)
	res := topic.Publish(ctx, &pubsub.Message{
		Data: msg,
	})

	id, err := res.Get(ctx)
	if err != nil {
		handleGoogleError(w, "publish message", err)
		return
	}

	bts, err := json.Marshal(publishMessageResponse{
		ProjectID: projectID,
		MessageID: id,
	})
	if err != nil {
		logWithPrefix("server: %+v", errors.Wrap(err, "could not encode publish result as JSON"))
		http.Error(w, "could not encode publish result as JSON", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, "publish_result.json", time.Time{}, bytes.NewReader(bts))
}

func (srv *Server) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectID := chi.URLParam(r, "projectID")
	topicID := chi.URLParam(r, "topicID")

	client, ok := srv.clients[projectID]
	if !ok {
		logWithPrefix("server: %+v", errors.Errorf("no client configured for project %q", projectID))
		http.Error(w, fmt.Sprintf("project %q not supported", projectID), http.StatusBadRequest)
		return
	}

	var req createSubscriptionRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "could not decode create topic request", http.StatusBadRequest)
		return
	}

	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		handleGoogleError(w, "check for topic existence", err)
		return
	}
	if !exists {
		http.Error(w, fmt.Sprintf("topic %q does not exist", topicID), http.StatusBadRequest)
		return
	}

	subscription, err := client.CreateSubscription(ctx, req.Name, pubsub.SubscriptionConfig{
		Topic: topic,
	})
	if err != nil {
		actionTried := fmt.Sprintf("create subscription %q on topic %q in project %q", req.Name, topicID, projectID)
		handleGoogleError(w, actionTried, err)
		return
	}

	bts, err := json.Marshal(createSubscriptionResponse{
		Subscription: struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			ProjectID string `json:"projectId"`
			TopicID   string `json:"topicId"`
		}{
			ID:        subscription.ID(),
			Name:      req.Name,
			ProjectID: projectID,
			TopicID:   topicID,
		},
	})
	if err != nil {
		logWithPrefix("server: %+v", errors.Wrap(err, "could not encode create topic response as JSON"))
		http.Error(w, "could not encode create topic response as JSON", http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, "topic.json", time.Time{}, bytes.NewReader(bts))
}

func (srv *Server) Subscribe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	projectID := chi.URLParam(r, "projectID")
	topicID := chi.URLParam(r, "topicID")

	client, ok := srv.clients[projectID]
	if !ok {
		logWithPrefix("server: %+v", errors.Errorf("no client configured for project %q", projectID))
		http.Error(w, "project %q not supported", http.StatusBadRequest)
		return
	}

	topic := client.Topic(topicID)
	exists, err := topic.Exists(ctx)
	if err != nil {
		handleGoogleError(w, "check for topic existence", err)
		return
	}
	if !exists {
		http.Error(w, fmt.Sprintf("topic %q does not exist", topicID), http.StatusBadRequest)
		return
	}

	topicName := topicNameFromTopicID(topicID)
	subName := fmt.Sprintf("%s_pubsubui_%s", topicName, shortuuid.New())

	sub, err := client.CreateSubscription(ctx, subName, pubsub.SubscriptionConfig{
		Topic: topic,
	})
	if err != nil {
		handleGoogleError(w, "create subscription", err)
		return
	}
	defer sub.Delete(context.Background())

	messageCh := make(chan *pubsub.Message)

	go srv.sse.Subscribe(w, r, messageCh)

	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		messageCh <- msg
	})
	if err != nil {
		handleGoogleError(w, "receive messages", err)
		return
	}

	close(messageCh)
}

func (srv *Server) Start(ctx context.Context, host string, port uint) error {
	r := chi.NewRouter()
	r.Get("/healthy", srv.Healthy)
	r.Get("/ready", srv.Ready)
	r.Get("/api/projects", srv.ListProjects)
	r.Post("/api/projects/{projectID}/topics", srv.CreateTopic)
	r.Get("/api/projects/{projectID}/topics", srv.ListTopics)
	r.Post("/api/projects/{projectID}/topics/{topicID}", srv.Publish)
	r.Get("/api/projects/{projectID}/topics/{topicID}", srv.Subscribe)
	r.Post("/api/projects/{projectID}/topics/{topicID}/subscriptions", srv.CreateSubscription)

	for _, cfgFn := range srv.additionalRouterConfigs {
		cfgFn(r)
	}

	addr := fmt.Sprintf("%s:%d", host, port)

	logWithPrefix("server: starting at %s", addr)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: r,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		},
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return httpServer.ListenAndServe()
	})
	g.Go(func() error {
		<-ctx.Done()
		logWithPrefix("server: shutting down")
		return httpServer.Shutdown(context.Background())
	})

	g.Wait()

	logWithPrefix("server: shut down")

	return nil
}
