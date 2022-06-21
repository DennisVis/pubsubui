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

export const api = {
  subscribe(projectId: string, topicId: string, onOpen: () => void, onMessage: (msg: string) => void): () => void {
    const source = new EventSource(`/api/projects/${projectId}/topics/${topicId}`)

    source.addEventListener('open', () => {
      onOpen()
    })

    source.addEventListener('message', event => {
      onMessage(event.data)
    })

    source.addEventListener('error', event => {
      console.error('event source error', event)
    })

    return () => {
      source.close()
    }
  },
}
