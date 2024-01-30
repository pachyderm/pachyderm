import {ModelDB} from '@jupyterlab/observables';
import {Contents, ServerConnection} from '@jupyterlab/services';
import {PartialJSONObject} from '@lumino/coreutils';
import {Signal, ISignal} from '@lumino/signaling';
import {DocumentRegistry} from '@jupyterlab/docregistry';
import {URLExt} from '@jupyterlab/coreutils';
import {requestAPI} from '../../handler';
import {Paging} from './paging';
import {MOUNT_BROWSER_PREFIX} from './mount';

// How many files to render in the FileBrowser UI per page
const MAX_NUM_CONTENTS_PAGE = 100;
// How many items to request per page
const PAGINATION_NUMBER = 400;
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
  private _cache: {
    key: string | null;
    now: number | null;
    contents: Contents.IModel[];
  };
  private _path: string;
  private _nameSuffix: string;
  // DOM Node ID of the FileBrowser
  private _id: string;
  // Triggers a cd event without changing the current path in the FileBrowser which forces a re-render of the FileBrowser
  private _rerenderFileBrowser: () => Promise<void>;
  // Updates the pagination UI after changes in page, contents, or maxPage have been made.
  private _pagingUpdate: () => void;
  // Previous search filter used last time get was called. Used to track if the current page needs to be reset
  // to zero.
  private _previousFilter: string | null;

  constructor(
    registry: DocumentRegistry,
    path: string,
    nameSuffix: string,
    id: string,
    _rerenderFileBrowser: () => Promise<void>,
    _pagingUpdate: () => void,
  ) {
    this._registry = registry;
    this._model = {page: 0, max_page: 0};
    this._cache = {key: null, now: null, contents: []};
    this._path = path;
    this._nameSuffix = nameSuffix;
    this._id = id;
    this._rerenderFileBrowser = _rerenderFileBrowser;
    this._pagingUpdate = _pagingUpdate;
    this._previousFilter = null;
  }

  get name(): string {
    return MOUNT_BROWSER_PREFIX + this._nameSuffix;
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
      const now = Date.now();
      this._cache = {key: localPath, now, contents: response.content};
      this._fetchNextPage(response, now, url, getOptions);
    }

    // Filter contents if filter defined
    const filter = this.getFilter();
    const contents = !filter
      ? this._cache.contents
      : this._cache.contents.filter((content: Contents.IModel) => {
          return content.name.toLowerCase().includes(filter);
        });

    // Reset page to 1 if the filter has changed and update the previousFilter for this check
    // on the next call of this method.
    if (filter !== this._previousFilter) {
      this._model.page = 1;
    }
    this._previousFilter = filter;

    // Update max page and pagination UI after contents have changed.
    this._model.max_page = Math.ceil(contents.length / MAX_NUM_CONTENTS_PAGE);
    this._pagingUpdate();

    return {
      ...shallowResponse,
      content: contents.slice(
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

      // Trigger a change directory event without a path change to force re-render the FileBrowser, then fetch the next page of results async.
      await this._rerenderFileBrowser();
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

  // Gets the user defined file name filter from the JupyterLab FilterBox.
  private getFilter(): string | null {
    const selector = `#${this._id} .jp-FileBrowser-filterBox .jp-FilterBox input`;
    const node: HTMLInputElement | null = document.querySelector(selector);
    return node?.value?.toLowerCase() || null;
  }

  private async _get(
    url: string,
    options: PartialJSONObject,
  ): Promise<Contents.IModel> {
    url += URLExt.objectToQueryString(options);
    return requestAPI<Contents.IModel>(url, 'GET');
  }
}
