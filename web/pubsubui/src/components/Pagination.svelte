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
  import Button, { Group, Icon, Label } from '@smui/button'
  import { pages, topics } from '../lib/topic/stores'
</script>

<div class="pagination-wrapper">
  <Group variant="unelevated">
    <Button on:click={topics.prevPage} variant="outlined" disabled={$topics.loading || $topics.page <= 1}>
      <Label>
        <Icon class="material-icons">
          navigate_before
        </Icon>
      </Label>
    </Button>

    {#each $pages as page}
      <Button
        on:click={() => topics.selectPage(page)}
        variant={$topics.page === page ? 'unelevated' : 'outlined'}
        disabled={!page || $topics.loading}
        class={!!page ? 'number' : ''}
      >
        <Label>{!!page ? page : '...'}</Label>
      </Button>
    {/each}

    <Button
      on:click={topics.nextPage}
      variant="outlined"
      disabled={$topics.loading || $topics.page >= $topics.totalPages}
    >
      <Label>
        <Icon class="material-icons">
          navigate_next
        </Icon>
      </Label>
    </Button>
  </Group>
</div>

<style>
  .pagination-wrapper {
    display: flex;
    flex-direction: row;
    justify-content: center;
    align-items: center;
  }
</style>
