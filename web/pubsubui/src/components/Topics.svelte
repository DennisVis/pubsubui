<!--
 Copyright 2022 Dennis Vis
 
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
 
     http://www.apache.org/licenses/LICENSE-2.0
 
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

<script lang="ts">
  import Accordion from '@smui-extra/accordion'
  import LayoutGrid, { Cell } from '@smui/layout-grid'
  import Paper, { Content as PaperContent, Title } from '@smui/paper'
  import Pagination from './Pagination.svelte'
  import Topic from './Topic.svelte'
  import { topics } from '../lib/topic/stores'

  const minPageSize = 3

  let publishing = false
  let height = 0

  $: {
    const padding = 300
    const itemHeight = 57
    const pageSize = Math.floor((height - padding) / itemHeight)

    if (pageSize < minPageSize) {
      topics.setPageSize(minPageSize)
    } else {
      topics.setPageSize(pageSize)
    }
  }
</script>

<svelte:window bind:innerHeight={height} />

<div>
  <LayoutGrid class="topics-wrapper">
    <Cell span={12}>
      <Paper class="topics">
        <Title>Topics</Title>

        <br />

        <PaperContent class="topics-inner">
          <Accordion>
            {#each $topics.topics as topic}
            <Topic
              publishing={publishing}
              topic={topic} on:publish={() => publishing = true}
              on:done={() => publishing = false}
            />
            {/each}
          </Accordion>
        </PaperContent>
      </Paper>
    </Cell>

    <Cell span={12}>
      <Paper class="pagination">
        <PaperContent>
          <Pagination />
        </PaperContent>
      </Paper>
    </Cell>
  </LayoutGrid>
</div>

<style>
  * :global(.topics-inner) {
    height: calc(100vh - 330px);
    overflow: auto;
  }
</style>
