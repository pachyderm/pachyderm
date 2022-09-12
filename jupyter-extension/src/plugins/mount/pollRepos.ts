import {ISignal, Signal} from '@lumino/signaling';
import {Poll} from '@lumino/polling';
import partition from 'lodash/partition';
import {requestAPI} from '../../handler';
import {AuthConfig, Branch, mountState, Repo} from './types';
import {ServerConnection} from '@jupyterlab/services';

export const MOUNTED_STATES: mountState[] = [
  'unmounting',
  'mounted',
  'mounting',
  'error',
];
export const UNMOUNTED_STATES: mountState[] = [
  'gone',
  'discovering',
  'unmounted',
];

export type ServerStatus = {
  code: number;
  message?: string;
};

export class PollRepos {
  constructor(name: string) {
    this.name = name;
  }
  readonly name: string;

  private _rawData: Repo[] = [];

  private _mounted: Repo[] = [];
  private _unmounted: Repo[] = [];
  private _status: ServerStatus = {code: 999, message: ''};
  private _config: AuthConfig = {
    pachd_address: '',
    cluster_status: 'INVALID',
  };

  private _mountedSignal = new Signal<this, Repo[]>(this);
  private _unmountedSignal = new Signal<this, Repo[]>(this);
  private _statusSignal = new Signal<this, ServerStatus>(this);
  private _configSignal = new Signal<this, AuthConfig>(this);

  private _dataPoll = new Poll({
    auto: true,
    factory: async () => this.getData(),
    frequency: {
      interval: 2000,
      backoff: true,
      max: 5000,
    },
  });

  get mounted(): Repo[] {
    return this._mounted;
  }

  set mounted(data: Repo[]) {
    if (data === this._mounted) {
      return;
    }
    this._mounted = data;
    this._mountedSignal.emit(data);
  }

  get unmounted(): Repo[] {
    return this._unmounted;
  }

  set unmounted(data: Repo[]) {
    if (data === this._unmounted) {
      return;
    }
    this._unmounted = data;
    this._unmountedSignal.emit(data);
  }

  get status(): ServerStatus {
    return this._status;
  }

  set status(status: ServerStatus) {
    if (JSON.stringify(status) === JSON.stringify(this._status)) {
      return;
    }

    this._status = status;
    this._statusSignal.emit(status);
  }

  get config(): AuthConfig {
    return this._config;
  }

  set config(config: AuthConfig) {
    if (JSON.stringify(config) === JSON.stringify(this._config)) {
      return;
    }
    this._config = config;
    this._configSignal.emit(config);
  }

  get mountedSignal(): ISignal<this, Repo[]> {
    return this._mountedSignal;
  }
  get unmountedSignal(): ISignal<this, Repo[]> {
    return this._unmountedSignal;
  }

  get statusSignal(): ISignal<this, ServerStatus> {
    return this._statusSignal;
  }

  get configSignal(): ISignal<this, AuthConfig> {
    return this._configSignal;
  }

  get poll(): Poll {
    return this._dataPoll;
  }

  refresh = async (): Promise<void> => {
    await this._dataPoll.refresh();
    await this._dataPoll.tick;
  };

  updateData = (data: Repo[]): void => {
    if (JSON.stringify(data) !== JSON.stringify(this._rawData)) {
      this._rawData = data;
      const [mountedPartition, unmountedPartition] = partition(
        data,
        (rep: Repo) => findMountedBranch(rep),
      );
      this.mounted = mountedPartition;
      this.unmounted = unmountedPartition;
    }
  };

  async getData(): Promise<void> {
    try {
      const config = await requestAPI<AuthConfig>('config', 'GET');
      this.config = config;
      if (config.cluster_status !== 'INVALID') {
        const data = await requestAPI<Repo[]>('repos', 'GET');
        this.status = {code: 200};
        this.updateData(data);
      }
    } catch (error) {
      if (error instanceof ServerConnection.ResponseError) {
        this.status = {
          code: error.response.status,
          message: error.response.statusText,
        };
      }
    }
  }
}

export function findMountedBranch(repo: Repo): Branch | undefined {
  //NOTE: Using find will cause issues if we allow multiple branches to be mounted at the same time.
  return repo.branches.find(
    (branch) => branch.mount[0].state && isMounted(branch.mount[0].state),
  );
}

export function isMounted(state: mountState): boolean {
  return MOUNTED_STATES.includes(state);
}
