import {ISignal, Signal} from '@lumino/signaling';
import {Poll} from '@lumino/polling';
import partition from 'lodash/partition';
import {requestAPI} from '../../handler';
import {Branch, mountState, Repo} from './types';

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

export class PollRepos {
  constructor(name: string) {
    this.name = name;
  }
  readonly name: string;

  private _rawData: Repo[] = [];

  private _mounted: Repo[] = [];
  private _unmounted: Repo[] = [];
  private _authenticated = false;

  private _mountedSignal = new Signal<this, Repo[]>(this);
  private _unmountedSignal = new Signal<this, Repo[]>(this);
  private _authenticatedSignal = new Signal<this, boolean>(this);

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

  get authenticated(): boolean {
    return this._authenticated;
  }

  set authenticated(authenticated: boolean) {
    if (authenticated === this._authenticated) {
      return;
    }
    this._authenticated = authenticated;
    this._authenticatedSignal.emit(authenticated);
  }

  get mountedSignal(): ISignal<this, Repo[]> {
    return this._mountedSignal;
  }
  get unmountedSignal(): ISignal<this, Repo[]> {
    return this._unmountedSignal;
  }

  get authenticatedSignal(): ISignal<this, boolean> {
    return this._authenticatedSignal;
  }

  get poll(): Poll {
    return this._dataPoll;
  }

  async getData(): Promise<void> {
    try {
      const data = await requestAPI<Repo[]>('repos', 'GET');
      this.authenticated = true;
      if (JSON.stringify(data) !== JSON.stringify(this._rawData)) {
        this._rawData = data;
        const [mountedPartition, unmountedPartition] = partition(
          data,
          (rep: Repo) => findMountedBranch(rep),
        );
        this.mounted = mountedPartition;
        this.unmounted = unmountedPartition;
      }
    } catch (error) {
      //TODO: add status to requestAPI
      this.authenticated = false;
    }
  }
}

export function findMountedBranch(repo: Repo): Branch | undefined {
  //NOTE: Using find will cause issues if we allow multiple branches to be mounted at the same time.
  return repo.branches.find(
    (branch) => branch.mount.state && isMounted(branch.mount.state),
  );
}

export function isMounted(state: mountState): boolean {
  return MOUNTED_STATES.includes(state);
}
