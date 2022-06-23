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
import { jsonToPubSubMessage } from './parse'
import { MessagesState } from './types'

function createMessages() {
  let unSub: () => void | undefined

  const { subscribe, update } = writable<MessagesState>(new MessagesState(false, false, '', []))

  function unSubscribe() {
    update(() => new MessagesState(false, false, '', []))

    if (unSub) {
      unSub()
    }
  }

  function connect(projectId: string, topicId: string) {
    unSubscribe()
    update(() => new MessagesState(true, false, topicId, []))

    unSub = api.subscribe(
      projectId,
      topicId,
      () => {
        update(s => new MessagesState(false, true, s.topic, s.messages))
      },
      message => update(s => {
        const newMessage = jsonToPubSubMessage(JSON.parse(message))
        return new MessagesState(s.connecting, s.open, s.topic, [newMessage, ...s.messages])
      }),
      err => update(() => new MessagesState(false, false, '', [], err)),
    )
  }

  return {
    subscribe,
    unSubscribe,
    connect,
  }
}

export const messages = createMessages()
