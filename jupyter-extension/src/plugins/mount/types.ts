import {SplitPanel} from '@lumino/widgets';

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
  authorization: authorization;
  branches: string[];
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
