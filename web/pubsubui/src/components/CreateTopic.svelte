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
  import { createEventDispatcher } from 'svelte'
  import Button, { Label } from '@smui/button'
  import Dialog, { Title, Content, Actions } from '@smui/dialog'
  import IconButton from '@smui/icon-button'
  import HelperText from '@smui/textfield/helper-text'
  import type { SnackbarComponentDev } from '@smui/snackbar'
  import Snackbar, { Actions as SnackActions, Label as SnackLabel } from '@smui/snackbar'
  import Textfield from '@smui/textfield'
  import { activeProject } from '../lib/project/stores'
  import { topics } from '../lib/topic/stores'

  const dispatch = createEventDispatcher()

  export let open: boolean = false

  let newTopicName: string | undefined
  let snackbarTopicCreationSucceeded: SnackbarComponentDev
  let snackbarTopicCreationFailed: SnackbarComponentDev

  function createTopic(projectId: string | undefined, topicName: string | undefined) {
    if (!projectId || !topicName) {
      return
    }

    dispatch('creating')

    try {
      topics.createTopic(projectId, topicName)
      snackbarTopicCreationSucceeded.open()
    } catch (err) {
      console.error('could not create topic', topicName, err)
      snackbarTopicCreationFailed.open()
    } finally {
      dispatch('done')
    }
  }

  function closeHandler() {
    newTopicName = undefined
    open = false
  }
</script>

<Dialog
  bind:open
  selection
  aria-labelledby="create-topic-title"
  aria-describedby="create-topic-content"
  on:SMUIDialog:closed={closeHandler}
>
  <Title>Create topic</Title>

  <Content id="create-topic-content">
    <div class="form-wrapper">
      <Textfield bind:value={newTopicName} input$emptyValueUndefined label="Name">
        <HelperText slot="helper">The name for the new topic</HelperText>
      </Textfield>
    </div>
  </Content>

  <Actions>
    <Button>
      <Label>Cancel</Label>
    </Button>

    <Button on:click={() => createTopic($activeProject, newTopicName)}>
      <Label>Create</Label>
    </Button>
  </Actions>
</Dialog>

<Snackbar bind:this={snackbarTopicCreationSucceeded}>
  <SnackLabel>Creating the topic "{newTopicName}" succeeded.</SnackLabel>
  <SnackActions>
    <IconButton class="material-icons" title="Dismiss">close</IconButton>
  </SnackActions>
</Snackbar>

<Snackbar bind:this={snackbarTopicCreationFailed}>
  <SnackLabel>Creating the topic "{newTopicName}" failed.</SnackLabel>
  <SnackActions>
    <IconButton class="material-icons" title="Dismiss">close</IconButton>
  </SnackActions>
</Snackbar>

<style>
  .form-wrapper {
    display: flex;
    flex-direction: column;
    padding: 15px 24px;
  }
</style>
