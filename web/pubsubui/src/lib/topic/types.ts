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

export class MessagePayload {
  constructor(
    readonly name: string,
    readonly payload: any,
  ) {}
}

export class Topic {
  constructor(
    readonly id: string,
    readonly name: string,
    readonly projectId: string,
    readonly payloads: MessagePayload[],
  ) {}
}

export class CreateTopicResponse {
  constructor(
    readonly topic: Topic,
  ){}
}

export class ListTopicsResponse {
  constructor(
    readonly projectId: string,
    readonly topics: Topic[],
    readonly totalItems: number,
    readonly page: number,
    readonly pageSize: number,
    readonly totalPages: number,
  ) {}
}

export class TopicsState {
  constructor(
    readonly loading: boolean,
    readonly topics: Topic[],
    readonly page: number,
    readonly totalPages: number,
  ){}
}

export class PublishMessageResponse {
  constructor(
    readonly projectId: string,
    readonly messageId: string,
  ) {}
}
