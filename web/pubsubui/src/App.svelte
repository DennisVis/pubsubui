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
  import { Icon } from '@smui/common'
  import { Svg } from '@smui/common/elements'
  import Fab from '@smui/fab'
  import IconButton from '@smui/icon-button'
  import LayoutGrid, { Cell } from '@smui/layout-grid'
  import Tooltip, { Wrapper } from '@smui/tooltip'
  import TopAppBar, { Row, Section, Title } from '@smui/top-app-bar'
  import IconPubSub from './assets/icons/IconPubSub.svelte'
  import CreateTopic from './components/CreateTopic.svelte'
  import Messages from './components/Messages.svelte'
  import Projects from './components/Projects.svelte'
  import Topics from './components/Topics.svelte'
  import { theme } from './lib/theme/stores'
  import { appWindow, windowHeight, windowWidth } from './lib/window/stores'

  let creatingTopic: boolean = false

  function toggleCreatingTopic() {
    creatingTopic = !creatingTopic
  }
</script>

<svelte:window bind:innerHeight={$windowHeight} bind:innerWidth={$windowWidth} />

<div>
  <TopAppBar variant="standard">
    <Row>
      <Section>
        <Icon component={Svg} class="logo">
          <IconPubSub />
        </Icon>

        <Title>Google Cloud Pub/Sub UI</Title>
      </Section>

      <Section>
        <Projects />
      </Section>
      
      <Section align="end" toolbar>
        <IconButton aria-label="Theme" on:click={theme.toggle}>
          <Icon class="material-icons">
            {$theme === 'light' ? 'dark_mode' : 'light_mode'}
          </Icon>
        </IconButton>
      </Section>
    </Row>
  </TopAppBar>

  <main>
    <LayoutGrid>
      <Cell span={$appWindow.isDesktop ? 7 : 12}>
        <Topics />
      </Cell>
      <Cell span={$appWindow.isDesktop ? 5 : 12}>
        <Messages />
      </Cell>
    </LayoutGrid>
  </main>

  <CreateTopic open={creatingTopic} />

  <Wrapper>
    <Fab color="secondary" on:click={toggleCreatingTopic} class="add-topic">
      <Icon class="material-icons">add</Icon>
    </Fab>

    <Tooltip yPos="above">Create new topic</Tooltip>
  </Wrapper>
</div>

<style>
  :global(body) {
    margin: 0;
  }

  main {
    max-height: calc(100vh - 64px);
    padding-top: 40px;
  }

  * :global(.logo) {
    height: 24px;
    width: 24px;
  }

  * :global(.add-topic) {
    bottom: 25px;
    position: fixed;
    right:25px;
  }
</style>
