import {ISignal, Signal} from '@lumino/signaling';
import {Poll} from '@lumino/polling';
import partition from 'lodash/partition';
import {requestAPI} from '../../handler';
import {Repo} from './mount';

export class PollRepos {
  constructor(name: string) {
    this.name = name;
  }
  readonly name: string;

  private _rawData: Repo[] = [];

  private _mounted: Repo[] = [];
  private _unmounted: Repo[] = [];

  private _mountedSignal = new Signal<this, Repo[]>(this);
  private _unmountedSignal = new Signal<this, Repo[]>(this);

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

  get mountedSignal(): ISignal<this, Repo[]> {
    return this._mountedSignal;
  }
  get unmountedSignal(): ISignal<this, Repo[]> {
    return this._unmountedSignal;
  }

  get poll(): Poll {
    return this._dataPoll;
  }

  async getData(): Promise<void> {
    const data = await requestAPI<Repo[]>('repos', 'GET');
    if (JSON.stringify(data) !== JSON.stringify(this._rawData)) {
      this._rawData = data;
      const [mountedPartition, unmountedPartition] = partition(
        data,
        (rep: Repo) =>
          rep.branches.find((branch) => branch.mount.state === 'mounted'),
      );
      this.mounted = mountedPartition;
      this.unmounted = unmountedPartition;
    }
  }
}
