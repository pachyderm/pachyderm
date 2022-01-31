import {SplitPanel} from '@lumino/widgets';

export type Branch = {
  branch: string;
  mount: Mount;
};

export type Mount = {
  name: string | null;
  state: string | null;
  mode: string | null;
};

export type Repo = {
  repo: string;
  branches: Branch[];
};

export interface IMountPlugin {
  mountedRepos: Repo[];
  unmountedRepos: Repo[];
  layout: SplitPanel;
}
