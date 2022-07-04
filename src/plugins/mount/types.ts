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

export type Branch = {
  branch: string;
  mount: Mount[];
};

export type Mount = {
  name: string;
  state: mountState;
  status: string;
  mode: string | null;
  mountpoint: string | null;
  mount_key: MountKey | null;
  how_many_commits_behind: number;
  actual_mounted_commit: string;
  latest_commit: string;
};

export type MountKey = {
  repo: string;
  branch: string;
  commit: string;
};

export type Repo = {
  repo: string;
  authorization: authorization;
  branches: Branch[];
};

export type AuthConfig = {
  cluster_status: clusterStatus;
  pachd_address?: string;
  server_cas?: string;
};

export interface IMountPlugin {
  mountedRepos: Repo[];
  unmountedRepos: Repo[];
  layout: SplitPanel;
  ready: Promise<void>;
}
