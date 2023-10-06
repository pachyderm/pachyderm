/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as GoogleProtobufAny from "../../google/protobuf/any.pb"

export enum State {
  STATE_UNKNOWN = "STATE_UNKNOWN",
  RUNNING = "RUNNING",
  SUCCESS = "SUCCESS",
  FAILURE = "FAILURE",
}

export type Group = {
}

export type Task = {
  id?: string
  state?: State
  input?: GoogleProtobufAny.Any
  output?: GoogleProtobufAny.Any
  reason?: string
  index?: string
}

export type Claim = {
}

export type TestTask = {
  id?: string
}