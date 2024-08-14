import {Contents} from '@jupyterlab/services';
import {TabPanel} from '@lumino/widgets';
import {JSONObject} from '@lumino/coreutils';

export type HealthCheckStatus =
  | 'UNHEALTHY'
  | 'HEALTHY_INVALID_CLUSTER'
  | 'HEALTHY_NO_AUTH'
  | 'HEALTHY_LOGGED_IN'
  | 'HEALTHY_LOGGED_OUT';

export type authorization = 'off' | 'none' | 'read' | 'write';

export type MountSettings = {
  defaultPipelineImage: string;
};

export type PfsInput = {
  pfs: {
    name?: string;
    repo: string;
    project?: string;
    branch?: string;
    commit?: string;
    files?: string[];
    glob: string;
    mode?: string;
  };
  cross?: JSONObject;
};

export type CrossInputSpec = {
  cross?: PfsInput[];
};

export type CurrentDatumResponse = {
  num_datums_received: number;
  input: {[key: string]: any};
  idx: number;
  all_datums_received: boolean;
};

export type DownloadPath = {
  path: string;
};

export type MountDatumResponse = {
  id: string;
  idx: number;
  num_datums_received: number;
  all_datums_received: boolean;
};

export type Repo = {
  name: string;
  project: string;
  uri: string;
  branches: Branch[];
};

export type Branch = {
  name: string;
  uri: string;
};

export type Repos = {
  [uri: string]: Repo;
};

export type MountedRepo = {
  repo: Repo;
  branch: Branch | null;
  commit: string | null;
};

export type Project = {
  name: string;
};

export type ProjectAuthInfo = {
  permissions: number[];
  roles: string[];
};

export type HealthCheck = {
  status: HealthCheckStatus;
  message?: string;
};

export type AuthConfig = {
  pachd_address?: string;
  server_cas?: string;
};

export interface IMountPlugin {
  repos: Repos;
  mountedRepo: MountedRepo | null;
  layout: TabPanel;
  ready: Promise<void>;
}

export type Pipeline = {
  name: string;
  project: Project | null;
};

export type PipelineSpec = {
  pipeline: Pipeline;
  description: string | null;
  transform: any;
  input: any;
  update: boolean;
  reprocess: boolean;
};

export type PpsMetadata = {
  version: string;
  config: PpsConfig;
};

export enum GpuMode {
  None = 'None',
  Simple = 'Simple',
  Advanced = 'Advanced',
}

// If this is updated, make sure to also update the corresponding `useEffect`
// call in ./components/Pipeline/hooks/usePipeline.tsx that writes this type to
// the notebook metadata.
export type PpsConfig = {
  pipeline: Pipeline;
  image: string;
  requirements: string | null;
  external_files: string | null;
  input_spec: string;
  port: string;
  gpu_mode: GpuMode;
  resource_spec: string | null;
};

export type PpsContext = {
  metadata: PpsMetadata | null;
  notebookModel: Omit<Contents.IModel, 'content'> | null;
};

export type CreatePipelineResponse = {
  message: string | null;
};

export interface IPachydermModel extends Contents.IModel {
  // The pachyderm uri for a file. This is not originally defined as a property in Contents.IModel, but
  // our endpoint for that model adds this.
  readonly file_uri: string;
}
