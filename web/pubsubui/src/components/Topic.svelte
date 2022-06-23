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
  import { JSONEditor } from 'svelte-jsoneditor'
  import { Content, Header, Panel } from '@smui-extra/accordion'
  import Button, { Icon, Label } from '@smui/button'
  import IconButton from '@smui/icon-button'
  import List, { Item, Text } from '@smui/list'
  import type { MenuComponentDev } from '@smui/menu'
  import Menu from '@smui/menu'
  import { Anchor } from '@smui/menu-surface'
  import type { SnackbarComponentDev } from '@smui/snackbar'
  import Snackbar, { Actions, Label as SnackLabel } from '@smui/snackbar'
  import CreateSubscription from './CreateSubscription.svelte'
  import { messages } from '../lib/message/stores'
  import { topics } from '../lib/topic/stores'
  import type { Topic } from '../lib/topic/types'
  import { theme } from '../lib/theme/stores'

  const dispatch = createEventDispatcher()

  export let publishing: boolean
  export let topic: Topic

  let panelOpen = false

  const defaultJsonContent = {
    text: '{}',
  }
  let jsonEditor: JSONEditor
  let content = defaultJsonContent

  let payloadMenu: MenuComponentDev
  let payloadMenuAnchor: HTMLDivElement
  let payloadMenuAnchorClasses: { [k: string]: boolean } = {}
  let snackbar: SnackbarComponentDev
  let snackbarMessage: string = ''

  let creatingSubscription: boolean = false

  function toggleCreatingSubscription() {
    creatingSubscription = !creatingSubscription
  }

  async function publishMessage() {
    snackbarMessage = ''

    dispatch('publish')

    try {
      await topics.publishMessage(topic.projectId, topic.id, content.text)
      snackbarMessage = `Published message to "${topic.name}".`
    } catch (err) {
      console.error('could not publish message', err)
      snackbarMessage = (err as Error).message
    } finally {
      snackbar.open()
      dispatch('done')
    }
  }

  $: {
    if ($topics.loading) {
      panelOpen = false
      content = defaultJsonContent
    }
  }
</script>

<div>
  <Snackbar bind:this={snackbar}>
    <SnackLabel>{snackbarMessage}</SnackLabel>
    <Actions>
      <IconButton class="material-icons" title="Dismiss">close</IconButton>
    </Actions>
  </Snackbar>

  <Panel bind:open={panelOpen} class={$theme === 'dark' ? 'jse-theme-dark' : ''} disabled={$topics.loading}>
    <Header>
      {topic.name}

      <IconButton slot="icon" toggle pressed={panelOpen}>
        <Icon class="material-icons" on>expand_less</Icon>
        <Icon class="material-icons">expand_more</Icon>
      </IconButton>
    </Header>

    <Content>
      <JSONEditor bind:this={jsonEditor} bind:content mode="code" />

      <br />

      <Button
        on:click={toggleCreatingSubscription}
        variant="unelevated"
        class="button-action button-shaped-round"
      >
        <Icon class="material-icons">add_circle</Icon>
        <Label>Create subscription</Label>
      </Button>

      <CreateSubscription open={creatingSubscription} topicId={topic.id} topicName={topic.name} />

      <Button
        on:click={() => messages.connect(topic.projectId, topic.id)}
        variant="unelevated"
        class="button-action button-shaped-round"
        disabled={$messages.connecting || $messages.open}
      >
        <Icon class="material-icons">move_to_inbox</Icon>
        <Label>Subscribe</Label>
      </Button>

      <Button
        on:click={() => publishMessage()}
        variant="unelevated"
        class="button-action button-shaped-round"
        disabled={publishing}
      >
        <Icon class="material-icons">send</Icon>
        <Label>Publish</Label>
      </Button>

      {#if topic.payloads.length > 0}
        <Button on:click={() => payloadMenu.setOpen(true)} class="button-payload">
          <Label>Select payload</Label>
        </Button>

        <div
          class={Object.keys(payloadMenuAnchorClasses).join(' ')}
          use:Anchor={{
            addClass: (className) => {
              if (!payloadMenuAnchorClasses[className]) {
                payloadMenuAnchorClasses[className] = true;
              }
            },
            removeClass: (className) => {
              if (payloadMenuAnchorClasses[className]) {
                delete payloadMenuAnchorClasses[className];
                payloadMenuAnchorClasses = payloadMenuAnchorClasses;
              }
            },
          }}
          bind:this={payloadMenuAnchor}
        >
          <Menu
            bind:this={payloadMenu}
            anchor={false}
            bind:anchorElement={payloadMenuAnchor}
            anchorCorner="BOTTOM_RIGHT"
          >
            <List>
              {#each topic.payloads as payload}
              <Item on:SMUI:action={() => {
                jsonEditor.set({
                  text: payload.payload,
                })
              }}>
                <Text>{payload.name}</Text>
              </Item>
              {/each}
            </List>
          </Menu>
        </div>
      {/if}
    </Content>
  </Panel>
</div>

<style>
  @import 'svelte-jsoneditor/themes/jse-theme-dark.css';

  * :global(.button-action) {
    margin-bottom: 5px;
  }

  * :global(.button-payload) {
    float: right;
  }

  * :global(.mdc-menu) {
    margin-left: -140px;
  }
</style>
