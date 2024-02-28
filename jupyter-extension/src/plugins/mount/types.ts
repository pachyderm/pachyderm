import {Contents} from '@jupyterlab/services';
import {TabPanel} from '@lumino/widgets';
import {JSONObject, ReadonlyJSONObject} from '@lumino/coreutils';

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

export type Mount = {
  name: string;
  repo: string;
  project: string;
  branch: string;
};

export type Repo = {
  repo: string;
  project: string;
  authorization: authorization;
  branches: string[];
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
  num_datums: number;
  input: {[key: string]: any};
  idx: number;
  all_datums_received: boolean;
};

export type MountDatumResponse = {
  id: string;
  idx: number;
  num_datums: number;
  all_datums_received: boolean;
};

export type ListMountsResponse = {
  mounted: {[key: string]: Mount};
  unmounted: {[key: string]: Repo};
};

export type Project = {
  name: string;
};

export type ProjectAuthInfo = {
  permissions: number[];
  roles: string[];
};

export type ProjectInfo = {
  project: Project;
  description: string;
  auth_info: ProjectAuthInfo;
  created_at: string;
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
  mountedRepos: Mount[];
  unmountedRepos: Repo[];
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
  notebookModel: Contents.IModel | null;
};

export type CreatePipelineResponse = {
  message: string | null;
};
