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

import { derived, writable } from 'svelte/store'
import { activeProject } from '../project/stores'
import { appWindow } from '../window/stores'
import { api } from './api'
import { Topic, TopicsState } from './types'

function createTopics() {
  let totalPages = 1
  const { subscribe, set, update } = writable<TopicsState>(new TopicsState(true, [], 1, 1))
  const page = writable<number>(1)
  const pageSize = writable<number>(17)

  const apiParams = derived(
    [activeProject, page, pageSize],
    ([$activeProject, $page, $pageSize]) => {
      return {
        projectId: $activeProject,
        page: $page,
        pageSize: $pageSize,
      }
    },
  )

  async function fetchTopics(projectId: string, page: number, pageSize: number) {
    update(s => {
      return new TopicsState(
        true,
        s.topics,
        s.page,
        s.totalPages,
      )
    })

    try {
      const ltr = await api.listTopics(projectId, page, pageSize)

      totalPages = ltr.totalPages

      set(new TopicsState(
        false,
        ltr.topics,
        ltr.page,
        totalPages,
      ))
    } catch (err) {
      console.error('could not call topics endpoint', err)
      update(s => new TopicsState(
        false,
        s.topics,
        s.page,
        s.totalPages,
      ))
      throw err
    }
  }

  async function createTopic(projectId: string, name: string) {
    update(s => {
      return new TopicsState(
        true,
        s.topics,
        s.page,
        s.totalPages,
      )
    })

    try {
     const res = await api.createTopic(projectId, name)

     update(s => new TopicsState(
        false,
        [res.topic, ...s.topics],
        s.page,
        s.totalPages,
     ))
    } catch (err) {
      console.error('could not create topic', err)
      update(s => new TopicsState(
        false,
        s.topics,
        s.page,
        s.totalPages,
      ))
      throw err
    }
  }

  async function publishMessage(projectId: string, topicId: string, message: string) {
    try {
      await api.publishMessage(projectId, topicId, message)
    } catch (err) {
      console.error('could not publish message', err)
      throw err
    }
  }

  function setPageSize(newPageSize: number) {
    pageSize.set(newPageSize)
    page.set(1)
  }

  function selectPage(newPage: number) {
    page.update(p => {
      if (newPage >= 1 && newPage <= totalPages) {
        return newPage
      }

      return p
    })
  }

  function prevPage() {
    page.update(p => {
      if (p > 1) {
        return p - 1
      }

      return p
    })
  }

  function nextPage() {
    page.update(p => {
      if (p < totalPages) {
        return p + 1
      }

      return p
    })
  }

  apiParams.subscribe(params => {
    if (!params.projectId) {
      return
    }

    fetchTopics(params.projectId as string, params.page, params.pageSize)
  })

  return {
    subscribe,
    setPageSize,
    selectPage,
    prevPage,
    nextPage,
    createTopic,
    publishMessage,
  }
}

export const topics = createTopics()

export const pages = derived([appWindow, topics], ([$appWindow, $topics]) => {
  const isBig = $appWindow.width >= 1160
  const numberOfSlots = isBig ? 7 : 4

  switch (true) {
    case $topics.totalPages <= numberOfSlots:
      return Array($topics.totalPages).fill(1).map((_, i) => i + 1)
    case isBig && $topics.page <= 4:
      return [1, 2, 3, 4, 5, 0, $topics.totalPages]
    case isBig && $topics.page >= $topics.totalPages - 3:
      return [1, 0, ...Array(5).fill(1).map((_, i) => $topics.totalPages - 4 + i)]
    case isBig && $topics.totalPages > numberOfSlots:
      return [1, 0, $topics.page - 1, $topics.page, $topics.page + 1, 0, $topics.totalPages]
    case $topics.totalPages > numberOfSlots && $topics.page == 1:
      return [1, 2, 0]
    case $topics.totalPages > numberOfSlots && $topics.page == $topics.totalPages:
      return [0, $topics.page - 1, $topics.page]
    case $topics.totalPages > numberOfSlots:
      return [0, $topics.page, 0]
    default:
      return []
  }
})
