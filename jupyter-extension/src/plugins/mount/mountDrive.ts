import {ModelDB} from '@jupyterlab/observables';
import {Contents, ServerConnection} from '@jupyterlab/services';
import {PartialJSONObject} from '@lumino/coreutils';
import {Signal, ISignal} from '@lumino/signaling';
import {DocumentRegistry} from '@jupyterlab/docregistry';
import {URLExt} from '@jupyterlab/coreutils';
import {requestAPI} from '../../handler';
import {Paging} from './paging';

const MAX_NUM_CONTENTS_PAGE = 100;
const DEFAULT_CONTENT_MODEL: Contents.IModel = {
  name: '',
  path: '',
  last_modified: '',
  created: '',
  content: [],
  format: 'json',
  mimetype: '',
  size: undefined,
  writable: true,
  type: 'directory',
};

export class MountDrive implements Contents.IDrive {
  public _registry: DocumentRegistry;
  readonly _model: Paging.IModel;
  private _fileChanged = new Signal<this, Contents.IChangedArgs>(this);
  private _loading = new Signal<this, boolean>(this);
  private _serverSettings = ServerConnection.makeSettings();
  private _isDisposed = false;
  private _cache: {key: string | null; contents: any};
  modelDBFactory?: ModelDB.IFactory | undefined;

  constructor(registry: DocumentRegistry) {
    this._registry = registry;
    this._model = {page: 0, max_page: 0};
    this._cache = {key: null, contents: DEFAULT_CONTENT_MODEL};
  }

  get name(): 'mount-browser' {
    return 'mount-browser';
  }

  get model(): Paging.IModel {
    return this._model;
  }

  get fileChanged(): ISignal<this, Contents.IChangedArgs> {
    return this._fileChanged;
  }

  // Signal emits `true` on expensive request to backend and `false` on response.
  get loading(): ISignal<this, boolean> {
    return this._loading;
  }

  get serverSettings(): ServerConnection.ISettings {
    return this._serverSettings;
  }
  get isDisposed(): boolean {
    return this._isDisposed;
  }

  async _get(
    url: string,
    options: PartialJSONObject,
  ): Promise<Contents.IModel> {
    url += URLExt.objectToQueryString(options);
    return requestAPI<Contents.IModel>(url, 'GET');
  }

  // We might be able to pass pagination to the ContentsManager backend by including
  // it in the URI specifier. This would require a custom ContentsManager which
  // supports pagination.
  async get(
    localPath: string,
    options?: Contents.IFetchOptions,
  ): Promise<Contents.IModel> {
    const url = URLExt.join('pfs', localPath);

    // Make a no-content request to determine the type of content being requested.
    // If the content type is not a directory, then we want to preserve our cache
    //   and not paginate the results.
    let shallowResponse;
    try {
      shallowResponse = await this._get(url, {content: '0'});
    } catch (e) {
      console.log('/pfs not found');
      return DEFAULT_CONTENT_MODEL;
    }
    const content = options?.content ? '1' : '0';
    if (content === '0') {
      return shallowResponse;
    }

    if (shallowResponse.type !== 'directory') {
      return await this._get(url, {...options, content});
    }

    // If we don't have contents cached, then we fetch them and cache the results.
    if (localPath !== this._cache.key || !localPath) {
      this._loading.emit(true);
      const response = await this._get(url, {...options, content});
      this._loading.emit(false);
      this._model.page = 1;
      this._model.max_page = Math.ceil(
        response.content.length / MAX_NUM_CONTENTS_PAGE,
      );
      this._cache = {key: localPath, contents: response.content};
    }

    return {
      ...shallowResponse,
      content: this._cache.contents.slice(
        (this._model.page - 1) * MAX_NUM_CONTENTS_PAGE,
        this._model.page * MAX_NUM_CONTENTS_PAGE,
      ),
    };
  }

  getDownloadUrl(localPath: string): Promise<string> {
    throw new Error('Method not implemented.');
  }
  newUntitled(options?: Contents.ICreateOptions): Promise<Contents.IModel> {
    throw new Error('Method not implemented.');
  }
  delete(localPath: string): Promise<void> {
    throw new Error('Method not implemented.');
  }
  rename(oldLocalPath: string, newLocalPath: string): Promise<Contents.IModel> {
    throw new Error('Method not implemented.');
  }
  save(
    localPath: string,
    options?: Partial<Contents.IModel>,
  ): Promise<Contents.IModel> {
    throw new Error('Method not implemented.');
  }
  copy(localPath: string, toLocalDir: string): Promise<Contents.IModel> {
    throw new Error('Method not implemented.');
  }
  createCheckpoint(localPath: string): Promise<Contents.ICheckpointModel> {
    return Promise.resolve({id: 'e', last_modified: 'e'});
  }
  listCheckpoints(localPath: string): Promise<Contents.ICheckpointModel[]> {
    return Promise.resolve([]);
  }
  restoreCheckpoint(localPath: string, checkpointID: string): Promise<void> {
    return Promise.resolve();
  }
  deleteCheckpoint(localPath: string, checkpointID: string): Promise<void> {
    return Promise.resolve();
  }

  dispose(): void {
    if (this.isDisposed) {
      return;
    }
    this._isDisposed = true;
    Signal.clearData(this);
  }
}
