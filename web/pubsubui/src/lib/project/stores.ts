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
import { ProjectsState } from './types'

function createActiveProject() {
  const { subscribe, set } = writable<string | undefined>()

  function setActiveProject(projectId: string | undefined) {
    if (!projectId) {
      return
    }

    set(projectId)
  }

  return {
    subscribe,
    set: setActiveProject,
  }
}

export const activeProject = createActiveProject()

function createProjects() {
  const { subscribe, set, update } = writable<ProjectsState>(new ProjectsState(true, []))

  async function fetchProjects() {
    update(s => new ProjectsState(true, s.projects))

    try {
      const lpr = await api.listProjects()

      set(new ProjectsState(false, lpr.projects))

      activeProject.set(lpr.projects[0])
    } catch (err) {
      console.error('could not fetch projects', err)
      update(s => new ProjectsState(false, s.projects))
    }
  }

  return {
    subscribe,
    fetchProjects,
  }
}

export const projects = createProjects()
