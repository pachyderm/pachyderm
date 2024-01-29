import {ModelDB} from '@jupyterlab/observables';
import {Contents, ServerConnection} from '@jupyterlab/services';
import {PartialJSONObject} from '@lumino/coreutils';
import {Signal, ISignal} from '@lumino/signaling';
import {DocumentRegistry} from '@jupyterlab/docregistry';
import {URLExt} from '@jupyterlab/coreutils';
import {requestAPI} from '../../handler';
import {Paging} from './paging';
import {MOUNT_BROWSER_PREFIX} from './mount';

const MAX_NUM_CONTENTS_PAGE = 100; // How many files to render in the FileBrowser UI per page
const PAGINATION_NUMBER = 400; // How many items to request per page
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
  _registry: DocumentRegistry;
  modelDBFactory?: ModelDB.IFactory | undefined;
  readonly _model: Paging.IModel;

  private _fileChanged = new Signal<this, Contents.IChangedArgs>(this);
  private _loading = new Signal<this, boolean>(this);
  private _serverSettings = ServerConnection.makeSettings();
  private _isDisposed = false;
  private _cache: {key: string | null; now: number | null; contents: any};
  private _path: string;
  private _name_suffix: string;
  private _refreshFileBrowser: () => Promise<void>; // A function that refresh the file browser

  constructor(
    registry: DocumentRegistry,
    path: string,
    name_suffix: string,
    refreshFileBrowser: () => Promise<void>,
  ) {
    this._registry = registry;
    this._model = {page: 0, max_page: 0};
    this._cache = {key: null, now: null, contents: DEFAULT_CONTENT_MODEL};
    this._path = path;
    this._name_suffix = name_suffix;
    this._refreshFileBrowser = refreshFileBrowser;
  }

  get name(): string {
    return MOUNT_BROWSER_PREFIX + this._name_suffix;
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
    const url = URLExt.join(this._path, localPath);

    // Make a no-content request to determine the type of content being requested.
    // If the content type is not a directory, then we want to preserve our cache
    //   and not paginate the results.
    let shallowResponse;
    try {
      shallowResponse = await this._get(url, {content: '0'});
    } catch (e) {
      console.log(url + ' not found');
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
      const getOptions = {
        ...options,
        number: PAGINATION_NUMBER,
        content,
      };
      const response = await this._get(url, getOptions);
      this._loading.emit(false);
      this._model.page = 1;
      this._model.max_page = Math.ceil(
        response.content.length / MAX_NUM_CONTENTS_PAGE,
      );
      const now = Date.now();
      this._cache = {key: localPath, now, contents: response.content};
      this._fetchNextPage(response, now, url, getOptions);
    }

    return {
      ...shallowResponse,
      content: this._cache.contents.slice(
        (this._model.page - 1) * MAX_NUM_CONTENTS_PAGE,
        this._model.page * MAX_NUM_CONTENTS_PAGE,
      ),
    };
  }

  _fetchNextPage(
    previousResponse: Contents.IModel,
    timeOfLastDirectoryChange: number,
    url: string,
    getOptions: PartialJSONObject,
  ): void {
    // Stop fetching pages if we recieve a page less than expected file count of a page
    if (previousResponse?.content?.length < PAGINATION_NUMBER) {
      return;
    }

    // Invoking an async function without awaiting it lets us start the process of fetching the next pages in the background while
    // continuing to render the the first page.
    (async () => {
      const nextResponse: Contents.IModel = await this._get(url, {
        ...getOptions,
        pagination_marker: previousResponse.content.slice(-1)[0].file_uri,
      });

      // Check to make sure that the time of the last actual directory change matches what is in the cache to prevent accidentally updating the
      // cache with results after the user has changed directories.
      if (this._cache.now !== timeOfLastDirectoryChange) {
        return;
      }

      // Update the cache contents and refresh the FileBrowser. Note this will not cause any changes in the current scroll position or file selection.
      // It will just add new pages and update the page selector without disrupting the user.
      this._cache.contents = this._cache.contents.concat(nextResponse.content);
      this._model.max_page = Math.ceil(
        this._cache.contents.length / MAX_NUM_CONTENTS_PAGE,
      );
      await this._refreshFileBrowser();
      this._fetchNextPage(
        nextResponse,
        timeOfLastDirectoryChange,
        url,
        getOptions,
      );
    })();
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
