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

import { writable } from 'svelte/store'
import { api } from './api'
import { SubscriptionState } from './types'

function createSubscriptions() {
  const { subscribe, update } = writable<SubscriptionState>(new SubscriptionState(true))

  async function createSubscription(
    projectId: string | undefined,
    topicId: string | undefined,
    subscriptionName: string | undefined,
  ) {
    update(s => new SubscriptionState(true))

    if (!projectId || !topicId || !subscriptionName) {
      return
    }

    try {
      await api.createSubscription(projectId, topicId, subscriptionName)
    } catch (err) {
      console.error('could not create subscription', subscriptionName, err)
      throw err
    } finally {
      update(s => new SubscriptionState(false))
    }
  }

  return {
    subscribe,
    createSubscription,
  }
}

export const subscriptions = createSubscriptions()