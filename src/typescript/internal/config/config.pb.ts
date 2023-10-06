/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

export enum ContextSource {
  NONE = "NONE",
  CONFIG_V1 = "CONFIG_V1",
  HUB = "HUB",
  IMPORTED = "IMPORTED",
}

export type Config = {
  userId?: string
  v1?: ConfigV1
  v2?: ConfigV2
}

export type ConfigV1 = {
  pachdAddress?: string
  serverCas?: string
  sessionToken?: string
  activeTransaction?: string
}

export type ConfigV2 = {
  activeContext?: string
  activeEnterpriseContext?: string
  contexts?: {[key: string]: Context}
  metrics?: boolean
  maxShellCompletions?: string
}

export type Context = {
  source?: ContextSource
  pachdAddress?: string
  serverCas?: string
  sessionToken?: string
  activeTransaction?: string
  clusterName?: string
  authInfo?: string
  namespace?: string
  portForwarders?: {[key: string]: number}
  clusterDeploymentId?: string
  enterpriseServer?: boolean
  project?: string
}