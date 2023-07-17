import {ISignal, Signal} from '@lumino/signaling';
import {Poll} from '@lumino/polling';
import {requestAPI} from '../../handler';
import {
  AuthConfig,
  Mount,
  ListMountsResponse,
  mountState,
  Repo,
  ProjectInfo,
} from './types';
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

export class PollMounts {
  constructor(name: string) {
    this.name = name;
  }
  readonly name: string;

  private _rawData: ListMountsResponse = <ListMountsResponse>{};

  private _mounted: Mount[] = [];
  private _unmounted: Repo[] = [];
  private _projects: ProjectInfo[] = [];
  private _status: ServerStatus = {code: 999, message: ''};
  private _config: AuthConfig = {
    pachd_address: '',
    cluster_status: 'INVALID',
  };

  private _mountedSignal = new Signal<this, Mount[]>(this);
  private _unmountedSignal = new Signal<this, Repo[]>(this);
  private _projectSignal = new Signal<this, ProjectInfo[]>(this);
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

  get mounted(): Mount[] {
    return this._mounted;
  }

  set mounted(data: Mount[]) {
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

  get projects(): ProjectInfo[] {
    return this._projects;
  }

  set projects(data: ProjectInfo[]) {
    if (data === this._projects) {
      return;
    }
    this._projects = data;
    this._projectSignal.emit(data);
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

  get mountedSignal(): ISignal<this, Mount[]> {
    return this._mountedSignal;
  }
  get unmountedSignal(): ISignal<this, Repo[]> {
    return this._unmountedSignal;
  }
  get projectSignal(): ISignal<this, ProjectInfo[]> {
    return this._projectSignal;
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

  updateData = (data: ListMountsResponse): void => {
    if (JSON.stringify(data) !== JSON.stringify(this._rawData)) {
      this._rawData = data;
      this.mounted = Array.from(Object.values(data.mounted));
      this.unmounted = Array.from(Object.values(data.unmounted));
    }
  };

  updateProjects = (data: ProjectInfo[]): void => {
    if (JSON.stringify(data) !== JSON.stringify(this.projects)) {
      this.projects = data;
    }
  };

  async getData(): Promise<void> {
    try {
      const config = await requestAPI<AuthConfig>('config', 'GET');
      this.config = config;
      if (config.cluster_status !== 'INVALID') {
        const data = await requestAPI<ListMountsResponse>('mounts', 'GET');
        this.status = {code: 200};
        this.updateData(data);
        const project = await requestAPI<ProjectInfo[]>('projects', 'GET');
        this.updateProjects(project);
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
