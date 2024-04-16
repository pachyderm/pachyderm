import {ISignal, Signal} from '@lumino/signaling';
import {Poll} from '@lumino/polling';
import {requestAPI} from '../../handler';
import {
  AuthConfig,
  HealthCheck,
  Repos,
  Repo,
  MountedRepo,
  Branch,
  CrossInputSpec,
  PfsInput,
} from './types';
import {ServerConnection} from '@jupyterlab/services';

export class PollMounts {
  static MOUNTED_REPO_LOCAL_STORAGE_KEY = 'mountedRepo';

  constructor(name: string) {
    this.name = name;

    const mountedRepoString = localStorage.getItem(
      PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY,
    );
    if (!mountedRepoString) {
      return;
    }

    try {
      const mountedRepo: MountedRepo = JSON.parse(mountedRepoString);
      this.mountedRepo = mountedRepo;
    } catch (e) {
      localStorage.removeItem(PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY);
    }
  }
  readonly name: string;

  private _repos: Repos = {};
  private _mountedRepo: MountedRepo | null = null;
  private _config: AuthConfig = {
    pachd_address: '',
  };
  private _health: HealthCheck = {
    status: 'HEALTHY_INVALID_CLUSTER',
    message: '',
  };

  private _reposSignal = new Signal<this, Repos>(this);
  private _mountedRepoSignal = new Signal<this, MountedRepo | null>(this);
  private _configSignal = new Signal<this, AuthConfig>(this);
  private _healthSignal = new Signal<this, HealthCheck>(this);

  private _dataPoll = new Poll({
    auto: true,
    factory: async () => this.getData(),
    frequency: {
      interval: 2000,
      backoff: true,
      max: 5000,
    },
  });

  get repos(): Repos {
    return this._repos;
  }

  set repos(data: Repos) {
    this._repos = data;
    this._reposSignal.emit(data);
  }

  get mountedRepo(): MountedRepo | null {
    return this._mountedRepo;
  }

  set mountedRepo(data: MountedRepo | null) {
    this._mountedRepo = data;
    this._mountedRepoSignal.emit(data);
  }

  get health(): HealthCheck {
    return this._health;
  }

  set health(healthCheck: HealthCheck) {
    if (JSON.stringify(healthCheck) === JSON.stringify(this._health)) {
      return;
    }

    this._health = healthCheck;
    this._healthSignal.emit(healthCheck);
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

  get reposSignal(): ISignal<this, Repos> {
    return this._reposSignal;
  }

  get mountedRepoSignal(): ISignal<this, MountedRepo | null> {
    return this._mountedRepoSignal;
  }

  get healthSignal(): ISignal<this, HealthCheck> {
    return this._healthSignal;
  }

  get configSignal(): ISignal<this, AuthConfig> {
    return this._configSignal;
  }

  get poll(): Poll {
    return this._dataPoll;
  }

  updateMountedRepo = (
    repo: Repo | null,
    mountedBranch: Branch | null,
  ): void => {
    if (repo === null) {
      localStorage.removeItem(PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY);
      this.mountedRepo = null;
      return;
    }

    if (!mountedBranch) {
      mountedBranch = repo?.branches[0] || null;
      for (const branch of repo.branches) {
        if (branch.name === 'master') {
          mountedBranch = branch;
        }
      }
    }

    this.mountedRepo = {
      mountedBranch,
      repo,
    };
    localStorage.setItem(
      PollMounts.MOUNTED_REPO_LOCAL_STORAGE_KEY,
      JSON.stringify(this.mountedRepo),
    );
  };

  getMountedRepoInputSpec = (): CrossInputSpec | PfsInput => {
    const mountedRepo = this.mountedRepo;
    if (mountedRepo === null) {
      return {};
    }

    let repo = mountedRepo.repo.name;
    if (mountedRepo.repo.project !== 'default') {
      repo = `${mountedRepo.repo.project}_name`;
    }

    return {
      pfs: {
        name: `${mountedRepo.repo.project}_${mountedRepo.repo.name}_${mountedRepo.mountedBranch.name}`,
        repo,
        glob: '/*',
      },
    };
  };

  refresh = async (): Promise<void> => {
    await this._dataPoll.refresh();
    await this._dataPoll.tick;
  };

  async getData(): Promise<void> {
    try {
      const healthCheck = await requestAPI<HealthCheck>('health', 'GET');
      this.health = healthCheck;
      if (
        healthCheck.status === 'HEALTHY_LOGGED_IN' ||
        healthCheck.status === 'HEALTHY_NO_AUTH'
      ) {
        const config = await requestAPI<AuthConfig>('config', 'GET');
        this.config = config;
        const data = await requestAPI<Repos>('repos', 'GET');
        this.repos = data;
      }
    } catch (error) {
      if (error instanceof ServerConnection.ResponseError) {
        this.health = {
          status: 'UNHEALTHY',
          message: error.response.statusText,
        };
      }
    }
  }
}
