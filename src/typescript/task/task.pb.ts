/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

export enum State {
  UNKNOWN = "UNKNOWN",
  RUNNING = "RUNNING",
  SUCCESS = "SUCCESS",
  FAILURE = "FAILURE",
  CLAIMED = "CLAIMED",
}

export type Group = {
  namespace?: string
  group?: string
}

export type TaskInfo = {
  id?: string
  group?: Group
  state?: State
  reason?: string
  inputType?: string
  inputData?: string
}

export type ListTaskRequest = {
  group?: Group
}