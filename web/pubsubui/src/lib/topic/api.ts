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

import { jsonToCreateTopicResponse, jsonToListTopicsResponse, jsonToPublishMessageResponse } from "./parse"
import type { CreateTopicResponse, ListTopicsResponse, PublishMessageResponse } from "./types"

export const api = {
  async createTopic(projectId: string, topicName: string): Promise<CreateTopicResponse> {
    try {
      const res = await fetch('/api/projects/' + projectId + '/topics', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: topicName,
        }),
      })
      if (res.status >= 400) {
        throw new Error(`could not create topic: ${await res.text()}`)
      }

      const json = await res.json()
      return jsonToCreateTopicResponse(json)
    } catch (err) {
      console.error('could not call create topic endpoint', err)
      throw err
    }
  },

  async listTopics(projectId: string, page: number, pageSize: number): Promise<ListTopicsResponse> {
    try {
      const res = await fetch(`/api/projects/${projectId}/topics?page=${page}&pageSize=${pageSize}`)
      if (res.status >= 400) {
        throw new Error(`could not list topics: ${await res.text()}`)
      }

      const json = await res.json()
      return jsonToListTopicsResponse(json)
    } catch (err) {
      console.error('could not call topics endpoint', err)
      throw err
    }
  },

  async publishMessage(projectId: string, topicId: string, message: any): Promise<PublishMessageResponse> {
    try {
      const res = await fetch(`/api/projects/${projectId}/topics/${topicId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: message,
      })
      if (res.status >= 400) {
        throw new Error(`could not publish message: ${await res.text()}`)
      }

      const json = await res.json()
      return jsonToPublishMessageResponse(json)
    } catch (err) {
      console.error('could not call publishMessage endpoint', err)
      throw err
    }
  },
}
