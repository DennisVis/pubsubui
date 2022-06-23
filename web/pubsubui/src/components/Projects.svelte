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
  import { onMount } from 'svelte'
  import Select, { Option } from '@smui/select'
  import { activeProject, projects } from '../lib/project/stores'
  import { messages } from '../lib/message/stores'
  import { topics } from '../lib/topic/stores'

  function handleProjectSelect(activeProject: string | undefined, newProject: string) {
    if (newProject === activeProject) {
      return
    }

    messages.unSubscribe()
    topics.selectPage(1)
  }

  onMount(() => {
    projects.fetchProjects()
  })
</script>

<div>
  <Select
    variant="outlined"
    bind:value={$activeProject}
    label="Project"
    disabled={$projects.loading}
    class="project-select"
  >
    {#each $projects.projects as project}
      <Option value={project} on:click={() => handleProjectSelect($activeProject, project)}>{project}</Option>
    {/each}
  </Select>
</div>

<style>
  * :global(.project-select.mdc-select--outlined:not(.mdc-select--disabled) .mdc-notched-outline__leading),
  * :global(.project-select.mdc-select--outlined:not(.mdc-select--disabled):not(.mdc-select--focused) .mdc-select__anchor:hover .mdc-notched-outline .mdc-notched-outline__leading),
  * :global(.project-select.mdc-select--outlined:not(.mdc-select--disabled) .mdc-notched-outline__notch),
  * :global(.project-select.mdc-select--outlined:not(.mdc-select--disabled):not(.mdc-select--focused) .mdc-select__anchor:hover .mdc-notched-outline .mdc-notched-outline__notch),
  * :global(.project-select.mdc-select--outlined:not(.mdc-select--disabled) .mdc-notched-outline__trailing),
  * :global(.project-select.mdc-select--outlined:not(.mdc-select--disabled):not(.mdc-select--focused) .mdc-select__anchor:hover .mdc-notched-outline .mdc-notched-outline__trailing) {
    border-color: #fff;
  }

  * :global(.project-select.mdc-select:not(.mdc-select--disabled) .mdc-floating-label),
  * :global(.project-select.mdc-select:not(.mdc-select--disabled) .mdc-select__selected-text) {
    color: #fff;
  }

  * :global(.project-select.mdc-select:not(.mdc-select--disabled) .mdc-select__dropdown-icon) {
    fill: #fff;
  }
</style>