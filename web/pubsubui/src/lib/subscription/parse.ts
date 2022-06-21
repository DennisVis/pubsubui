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

import { CreateSubscriptionResponse, Subscription } from "./types"

export function jsonToSubscription(json: any): Subscription {
  if (typeof(json.id) !== 'string') {
    throw new Error('ID in subscription JSON not a string')
  }
  if (typeof(json.name) !== 'string') {
    throw new Error('name in subscription JSON not a string')
  }
  if (typeof(json.projectId) !== 'string') {
    throw new Error('project ID in subscription JSON not a string')
  }
  if (typeof(json.topicId) !== 'string') {
    throw new Error('topic ID in subscription JSON not a string')
  }

  return new Subscription(
    json.id,
    json.name,
    json.projectId,
    json.topicId,
  )
}

export function jsonToCreateSubscriptionResponse(json: any): CreateSubscriptionResponse {
  if (typeof(json) !== 'object') {
    throw new Error('create subscription response JSON not an object')
  }
  if (typeof(json.subscription) !== 'object') {
    throw new Error('create subscription response JSON subscription not an object')
  }

  return new CreateSubscriptionResponse(
    jsonToSubscription(json.subscription),
  )
}
