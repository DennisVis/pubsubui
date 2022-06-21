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

import { jsonToListProjectsResponse } from './parse'
import type { ListProjectsResponse } from './types'

export const api = {
    async listProjects(): Promise<ListProjectsResponse> {
      try {
        const res = await fetch('/api/projects')
        const json = await res.json()
        return jsonToListProjectsResponse(json)
      } catch (err) {
        console.error('could not call projects endpoint', err)
        throw err
      }
    },
}