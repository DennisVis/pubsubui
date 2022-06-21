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
  import type { SnackbarComponentDev } from '@smui/snackbar'
  import Snackbar, { Actions as SnackActions, Label as SnackLabel } from '@smui/snackbar'
  import HelperText from '@smui/textfield/helper-text'
  import Textfield from '@smui/textfield'
  import { activeProject } from '../lib/project/stores'
  import { subscriptions } from '../lib/subscription/store'

  const dispatch = createEventDispatcher()

  export let open: boolean = false
  export let topicId: string | undefined
  export let topicName: string | undefined

  let newSubscriptionName: string | undefined
  let snackbarSubscriptionCreationSucceeded: SnackbarComponentDev
  let snackbarSubscriptionCreationFailed: SnackbarComponentDev

  function createSubscription(
    projectId: string | undefined,
    topicId: string | undefined,
    subcriptionName: string | undefined,
  ) {
    if (!projectId || !topicId || !subcriptionName) {
      return
    }

    dispatch('creating')

    try {
      subscriptions.createSubscription(projectId, topicId, subcriptionName)
      snackbarSubscriptionCreationSucceeded.open()
    } catch (err) {
      console.error('could not create subscription', subcriptionName, err)
      snackbarSubscriptionCreationFailed.open()
    } finally {
      dispatch('done')
    }
  }

  function closeHandler() {
    newSubscriptionName = undefined
    open = false
  }
</script>

<Dialog
  bind:open
  selection
  aria-labelledby="create-subscription-title"
  aria-describedby="create-subscription-content"
  on:SMUIDialog:closed={closeHandler}
>
  <Title>Create subscription for "{topicName}"</Title>

  <Content id="create-subscription-content">
    <div class="form-wrapper">
      <Textfield bind:value={newSubscriptionName} input$emptyValueUndefined label="Name">
        <HelperText slot="helper">The name for the new subscription</HelperText>
      </Textfield>
    </div>
  </Content>

  <Actions>
    <Button>
      <Label>Cancel</Label>
    </Button>

    <Button on:click={() => createSubscription($activeProject, topicId, newSubscriptionName)}>
      <Label>Create</Label>
    </Button>
  </Actions>
</Dialog>

<Snackbar bind:this={snackbarSubscriptionCreationSucceeded}>
  <SnackLabel>Creating the subscription "{newSubscriptionName}" on topic "{topicName}" succeeded.</SnackLabel>
  <SnackActions>
    <IconButton class="material-icons" title="Dismiss">close</IconButton>
  </SnackActions>
</Snackbar>

<Snackbar bind:this={snackbarSubscriptionCreationFailed}>
  <SnackLabel>Creating the subscription "{newSubscriptionName}" on topic "{topicName}" failed.</SnackLabel>
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
