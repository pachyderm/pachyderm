import {SplitPanel} from '@lumino/widgets';
import {JSONObject} from '@lumino/coreutils';

export type mountState =
  | 'unmounting'
  | 'mounted'
  | 'mounting'
  | 'error'
  | 'gone'
  | 'discovering'
  | 'unmounted'
  | '';

export type clusterStatus = 'INVALID' | 'AUTH_DISABLED' | 'AUTH_ENABLED';

export type authorization = 'off' | 'none' | 'read' | 'write';

export type Mount = {
  name: string;
  repo: string;
  project: string;
  branch: string;
  commit: string | null;
  glob: string | null;
  mode: string | null;
  state: mountState;
  status: string;
  mountpoint: string | null;
  how_many_commits_behind: number;
  actual_mounted_commit: string;
  latest_commit: string;
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
  curr_idx: number;
};

export type MountDatumResponse = {
  id: string;
  idx: number;
  num_datums: number;
};

export type ListMountsResponse = {
  mounted: {[key: string]: Mount};
  unmounted: {[key: string]: Repo};
};

export type AuthConfig = {
  cluster_status: clusterStatus;
  pachd_address?: string;
  server_cas?: string;
};

export interface IMountPlugin {
  mountedRepos: Mount[];
  unmountedRepos: Repo[];
  layout: SplitPanel;
  ready: Promise<void>;
}

export type SameMetadata = {
  apiVersion: string;
  environments: SameEnv;
  metadata: SameMetaMetadata;
  notebook: SameNotebookMetadata;
  run: SameRunMetadata;
};

export type SameEnv = {
  default: DefaultSameEnv;
};

export type DefaultSameEnv = {
  image_tag: string;
};
export type SameMetaMetadata = {
  labels?: string[];
  name: string;
  version?: string;
};

export type SameNotebookMetadata = {
  // Note: name and path are filled in when you pass the notebook to SAME
  requirements: string;
};

export type SameRunMetadata = {
  name: string;
  input?: string; //Note: SAME doesn't actually read this field when reading from the notebook and instead expects you to pass it on the command line
};

export type CreatePipelineResponse = {
  message: string | null;
};
