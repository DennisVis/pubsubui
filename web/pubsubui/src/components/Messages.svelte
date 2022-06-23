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
  import Button, { Icon, Label } from '@smui/button'
  import Card from '@smui/card'
  import IconButton from '@smui/icon-button'
  import LayoutGrid, { Cell } from '@smui/layout-grid'
  import Paper, { Content, Title } from '@smui/paper'
  import type { SnackbarComponentDev } from '@smui/snackbar'
  import Snackbar, { Actions as SnackActions, Label as SnackLabel } from '@smui/snackbar'
  import { messages } from '../lib/message/stores'
  import { theme } from '../lib/theme/stores'
  import { appWindow } from '../lib/window/stores'

  let titleSpan: number = 7
  let unsubSpan: number = 5

  let snackbar: SnackbarComponentDev
  let snackbarMessage: string = ''

  $: {
    switch (true) {
      case $appWindow.isTablet:
        titleSpan = 5
        unsubSpan = 3
        break
      case $appWindow.isPhone:
        titleSpan = 12
        unsubSpan = 12
        break
    }
  }

  $: {
    if (!!$messages.error) {
      snackbarMessage = $messages.error
      snackbar.open()
    } else {
      snackbarMessage = ''
      !!snackbar && snackbar.close()
    }
  }
</script>

<div>
  <LayoutGrid>
    <Cell span={12}>
      <Paper class="messages-paper">
        <Title>
          <LayoutGrid>
            <Cell span={titleSpan}>
              { $messages.connecting || $messages.open ? 'Subscribed to "' + $messages.topic + '"' : 'Not subscribed' }
            </Cell>

            <Cell span={unsubSpan}>
              <Button
                on:click={messages.unSubscribe}
                variant="unelevated"
                class="button-shaped-round button-unsubscribe"
                disabled={$messages.connecting || !$messages.open}
              >
                <Icon class="material-icons">unsubscribe</Icon>
                <Label>Unsubscribe</Label>
              </Button>
            </Cell>
          </LayoutGrid>
        </Title>

        <hr />

        <Content class="messages">
          {#each $messages.messages as message}
            <Card class="message">
              <table class="meta">
                <tr>
                  <td>
                    <strong>ID</strong>
                  </td>
                  <td>
                    {message.id}
                  </td>
                </tr>
                <tr>
                  <td>
                    <strong>Publish time</strong>
                  </td>
                  <td>
                    {message.publishDate.toLocaleDateString()}
                    {message.publishDate.toLocaleTimeString()}
                  </td>
                </tr>
                {#if !!message.attributes}
                <tr>
                  <td>
                    <strong>Attributes</strong>
                  </td>
                  <td>
                    <table>
                      {#each Object.entries(message.attributes || {}) as [key, val]}
                        <tr>
                          <td>{key}</td>
                          <td>=</td>
                          <td>{val}</td>
                        </tr>
                      {/each}
                    </table>
                  </td>
                </tr>
                {/if}
              </table>

              <div class={$theme === 'dark' ? 'code dark' : 'code'}>
                <pre><code>{JSON.stringify(message.data, null, 2)}</code></pre>
              </div>
            </Card>
          {/each}
        </Content>
      </Paper>
    </Cell>
  </LayoutGrid>
</div>

<Snackbar bind:this={snackbar}>
  <SnackLabel>{snackbarMessage}</SnackLabel>
  <SnackActions>
    <IconButton class="material-icons" title="Dismiss">close</IconButton>
  </SnackActions>
</Snackbar>

<style>
  * :global(.messages-paper) {
    height: calc(100vh - 166px);
  }

  * :global(.button-unsubscribe) {
    padding-right: 15px;
  }

  * :global(.messages) {
    height: calc(100% - 185px);
    overflow-y: auto;
    padding: 0 2px;
  }

  * :global(.message) {
    margin-bottom: 10px;
    
    padding: 1em;
  }

  * :global(.message .code) {
    background-color: #f0f0f0;
    overflow-x: auto;
    padding: 5px 10px;
  }

  * :global(.message .code.dark) {
    background-color: #343434;
  }

  * :global(.message .code pre) {
    margin: 0;
  }
</style>
