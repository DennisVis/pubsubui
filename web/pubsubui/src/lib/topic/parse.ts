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

import { CreateTopicResponse, ListTopicsResponse, MessagePayload, PublishMessageResponse, Topic } from './types'

function jsonToMessagePayload(json: any): MessagePayload {
  if (typeof(json) !== 'object') {
    throw new Error('message payload JSON not an object')
  }
  if (typeof(json.name) !== 'string') {
    throw new Error('message payload JSON did not contain a name string')
  }
  if (typeof(json.payload) !== 'string') {
    throw new Error('message payload JSON did not contain a payload string')
  }

  return new MessagePayload(json.name, json.payload)
}

export function jsonToTopic(json: any): Topic {
  if (typeof(json.id) !== 'string') {
    throw new Error('ID in topic JSON not a string')
  }
  if (typeof(json.name) !== 'string') {
    throw new Error('name in topic JSON not a string')
  }
  if (typeof(json.projectId) !== 'string') {
    throw new Error('project ID in topic JSON not a string')
  }

  const payloads = (json.payloads || []).map(jsonToMessagePayload)

  return new Topic(
    json.id,
    json.name,
    json.projectId,
    payloads,
  )
}

export function jsonToCreateTopicResponse(json: any): CreateTopicResponse {
  if (typeof(json) !== 'object') {
    throw new Error('create topic response JSON not an object')
  }
  if (typeof(json.topic) !== 'object') {
    throw new Error('create topic response JSON did not contain a topic object')
  }

  return new CreateTopicResponse(jsonToTopic(json.topic))
}

export function jsonToListTopicsResponse(json: any): ListTopicsResponse {
  if (typeof(json) !== 'object') {
    throw new Error('list topics response JSON not an object')
  }
  if (typeof(json.projectId) !== 'string') {
    throw new Error('list topics response JSON project ID not a string')
  }
  if (!Array.isArray(json.topics)) {
    throw new Error('list topics response JSON did not contain a topics array')
  }
  if (typeof(json.totalItems) !== 'number') {
    throw new Error('list topics response JSON total items not a number')
  }
  if (typeof(json.page) !== 'number') {
    throw new Error('list topics response JSON page not a number')
  }
  if (typeof(json.pageSize) !== 'number') {
    throw new Error('list topics response JSON page size not a number')
  }
  if (typeof(json.totalPages) !== 'number') {
    throw new Error('list topics response JSON total pages not a number')
  }

  const topics = json.topics.map(jsonToTopic)

  return new ListTopicsResponse(
    json.projectId,
    topics,
    json.totalItems,
    json.page,
    json.pageSize,
    json.totalPages,
  )
}

export function jsonToPublishMessageResponse(json: any): PublishMessageResponse {
  if (typeof(json) !== 'object') {
    throw new Error('publish message response JSON not an object')
  }
  if (typeof(json.projectId) !== 'string') {
    throw new Error('publish message response JSON project ID not a string')
  }
  if (typeof(json.messageId) !== 'string') {
    throw new Error('publish message response JSON message ID not a string')
  }

  return new PublishMessageResponse(
    json.projectId,
    json.messageId,
  )
}
