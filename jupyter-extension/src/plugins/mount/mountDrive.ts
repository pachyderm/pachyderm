import {ModelDB} from '@jupyterlab/observables';
import {Contents, ServerConnection} from '@jupyterlab/services';
import {PartialJSONObject} from '@lumino/coreutils';
import {Signal, ISignal} from '@lumino/signaling';
import {DocumentRegistry} from '@jupyterlab/docregistry';
import {URLExt} from '@jupyterlab/coreutils';
import {requestAPI} from '../../handler';
import {MOUNT_BROWSER_PREFIX} from './mount';

// How many Content li are visible at a given time for the user to scroll.
const VISIBLE_CONTENT_LI_COUNT = 500;
// How many items to request per page.
const PAGINATION_NUMBER = 1000;
// An empty default representation of directory.
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
  // Properties necessary for implementing IDrive, but otherwise unused.
  readonly _registry: DocumentRegistry;
  modelDBFactory?: ModelDB.IFactory | undefined;
  private _fileChanged = new Signal<this, Contents.IChangedArgs>(this);
  private _serverSettings = ServerConnection.makeSettings();
  private _isDisposed = false;

  // Signal emits `true` on navigating to a new directory and `false` on response.
  private _loading = new Signal<this, boolean>(this);
  // Contents cache for the Drive used to avoid the small Contents limit that can be visible in a FileBrowser.
  private _cache: {
    key: string | null;
    now: number | null;
    contents: Contents.IModel[];
    filteredContents: Contents.IModel[];
    shallowResponse: Contents.IModel;
  };
  // Root path of this Drive
  private _path: string;
  // Name suffix for this Drive. Must be unique across all Drive instances.
  private _nameSuffix: string;
  // DOM Node ID of the FileBrowser using this Drive.
  private _id: string;
  // Forces a re-render of the FileBrowser using this Drive.
  private _rerenderFileBrowser: () => Promise<void>;
  // Previous search filter value as of the previous _get call.
  private _previousFilter: string | null;
  // The index determining which set of cached Contents is visible to the user. Changed by scrolling the Contents DOM node.
  private _index: number;
  // True if the FileBrowser contents scrolling event listener has been setup, false if not. Avoids setting up multiple
  // scroll event listeners.
  private _hasScrollEventListener: boolean;
  // Function to call when the repo needs to be unmounted likely due to some error state.
  private _unmountRepo: () => void;

  constructor(
    registry: DocumentRegistry,
    path: string,
    nameSuffix: string,
    id: string,
    rerenderFileBrowser: () => Promise<void>,
    unmountRepo: () => void,
  ) {
    this._registry = registry;
    this._cache = {
      key: null,
      now: null,
      shallowResponse: DEFAULT_CONTENT_MODEL,
      contents: [],
      filteredContents: [],
    };
    this._path = path;
    this._nameSuffix = nameSuffix;
    this._id = id;
    this._rerenderFileBrowser = rerenderFileBrowser;
    this._previousFilter = null;
    this._index = 0;
    this._hasScrollEventListener = false;
    this._unmountRepo = unmountRepo;
  }

  get name(): string {
    return MOUNT_BROWSER_PREFIX + this._nameSuffix;
  }

  get fileChanged(): ISignal<this, Contents.IChangedArgs> {
    return this._fileChanged;
  }

  // Signal emits `true` on navigating to a new directory and `false` on response
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
    this.setupScrollingHandler();

    const url = URLExt.join(this._path, localPath);

    // If we have cached content return that
    if (localPath === this._cache.key && localPath) {
      return this._getCachedContent();
    }

    // Reset cache. It is important to reset this on any change in directory immediately. Otherwise the loading of files
    // that happens in the background for folders with a large amount of files might attempt to continue which could fail if the user
    // mounts a new repository.
    const now = Date.now();
    this._cache = {
      key: localPath,
      now,
      contents: [],
      filteredContents: [],
      shallowResponse: DEFAULT_CONTENT_MODEL,
    };

    // Make a no-content request to determine the type of content being requested.
    // If the content type is not a directory, then we want to preserve our cache
    //   and not paginate the results.
    let shallowResponse;
    try {
      shallowResponse = await this._get(url, {
        content: '0',
      });
    } catch (e) {
      console.debug('Get Error', url + ' not found. Unmounting repo.');
      this._unmountRepo();
      return DEFAULT_CONTENT_MODEL;
    }

    this._cache.shallowResponse = shallowResponse;

    const content = options?.content ? '1' : '0';
    if (content === '0') {
      this._loading.emit(false);
      return shallowResponse;
    }

    if (shallowResponse.type !== 'directory') {
      this._loading.emit(false);
      return await this._get(url, {...options, content});
    }

    // If we don't have contents cached, then we fetch them and cache the results.
    this.resetContentsNode();
    this._loading.emit(false);
    const getOptions = {
      ...options,
      number: PAGINATION_NUMBER,
      content,
    };
    this._loading.emit(true);
    await this._fetchNextPage(null, now, url, getOptions);
    return this._getCachedContent();
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

  private _getCachedContent(): Contents.IModel {
    const newFilter = this.getFilter();
    if (newFilter !== this._previousFilter) {
      this.resetContentsNode();
      this.filterContents();
    }
    this._previousFilter = newFilter;

    return {
      ...this._cache.shallowResponse,
      content: this._cache.filteredContents.slice(
        this._index,
        this._index + VISIBLE_CONTENT_LI_COUNT,
      ),
    };
  }

  private async _fetchNextPage(
    previousResponse: Contents.IModel | null,
    timeOfLastDirectoryChange: number,
    url: string,
    getOptions: PartialJSONObject,
  ): Promise<void> {
    const nextResponseParams = {
      ...getOptions,
    };
    if (previousResponse) {
      nextResponseParams.pagination_marker =
        previousResponse.content.slice(-1)[0].file_uri;
    }
    const nextResponse: Contents.IModel = await this._get(
      url,
      nextResponseParams,
    );

    // Check to make sure that the time of the last actual directory change matches what is in the cache to prevent accidentally updating the
    // cache with results after the user has changed directories.
    if (this._cache.now !== timeOfLastDirectoryChange) {
      this._loading.emit(false);
      return;
    }

    // Update the cache contents and refresh the FileBrowser. Note this will not cause any changes in the current scroll position or file selection.
    // It will just add new pages and update the page selector without disrupting the user.
    this._cache.contents = this._cache.contents.concat(nextResponse.content);
    this.filterContents();
    if (previousResponse) {
      await this.updateScrollPosition();
    }

    // Stop fetching pages if we recieve a page less than expected file count of a page
    if (nextResponse?.content?.length < PAGINATION_NUMBER) {
      this._loading.emit(false);
      return;
    }

    // Invoking an async function without awaiting it lets us start the process of fetching the next pages in the background while
    // continuing to render the the first page.
    this._fetchNextPage(
      nextResponse,
      timeOfLastDirectoryChange,
      url,
      getOptions,
    ).catch((e) => {
      // This can happen if a user unmounts the current repository in the middle of loading results.
      if (this._cache.now !== timeOfLastDirectoryChange) {
        this._loading.emit(false);
        return;
      }

      // This should never happen and means some critical backend error has occured.
      console.debug(`Failed Fetching Next Results ${e}. Umounting repo`);
      this._unmountRepo();
    });
  }

  private async updateScrollPosition(): Promise<void> {
    const contentsNode = this.getContentsNode();
    if (!contentsNode) {
      return;
    }

    if (contentsNode.scrollHeight === 0) {
      return;
    }

    const scrolledToBottom =
      Math.abs(
        contentsNode.scrollHeight -
          contentsNode.scrollTop -
          contentsNode.clientHeight,
      ) <= 1;
    if (!scrolledToBottom) {
      return;
    }
    const indexPercent = this._index / this._cache.filteredContents.length;
    contentsNode.scrollTop = contentsNode.scrollHeight * indexPercent;

    await this._rerenderFileBrowser();
  }

  // Gets the user defined file name filter from the JupyterLab FilterBox.
  private getFilter(): string | null {
    const selector = `#${this._id} .jp-FileBrowser-filterBox .jp-FilterBox input`;
    const node: HTMLInputElement | null = document.querySelector(selector);
    return node?.value?.toLowerCase() || null;
  }

  // Gets the DOM node containing the Contents of the FileBrowser.
  private getContentsNode(): HTMLStyleElement | null {
    const selector = `#${this._id} .jp-DirListing .jp-DirListing-content`;
    const node: HTMLStyleElement | null = document.querySelector(selector);
    return node;
  }

  // Scrolls the Contents DOM node to the top and resets the index. Should be called after a change in filter or directory.
  private resetContentsNode(): void {
    const contentsNode = this.getContentsNode();
    this._index = 0;
    if (contentsNode) {
      (contentsNode as any).scrollTop = 0;
    }
  }

  // Filters the cached contents into filteredContents. Should be called when the cached contents or filter changes
  private filterContents(): void {
    const filter = this.getFilter();
    this._cache.filteredContents = !filter
      ? this._cache.contents
      : this._cache.contents.filter((content: Contents.IModel) => {
          return content.name.toLowerCase().includes(filter);
        });
  }

  // Sets up the Contents scroll event listener necessary for infinite scroll pagination. Uses _hasScrollEventListener to ensure
  // the event listener is only setup once.
  private setupScrollingHandler(): void {
    if (this._hasScrollEventListener) {
      return;
    }

    const contentsNode = this.getContentsNode();
    if (!contentsNode) {
      return;
    }

    this._hasScrollEventListener = true;
    contentsNode.addEventListener('scroll', () => {
      const indexPercent = contentsNode.scrollTop / contentsNode.scrollHeight;
      // Prevent scrolling the index past the upper limit of visible content li nodes
      let index = Math.round(
        indexPercent * this._cache.filteredContents.length,
      );
      if (
        index + VISIBLE_CONTENT_LI_COUNT >
        this._cache.filteredContents.length
      ) {
        index = this._cache.filteredContents.length - VISIBLE_CONTENT_LI_COUNT;
      }
      this._index = index;
      this._rerenderFileBrowser();
    });
  }

  private async _get(
    url: string,
    options: PartialJSONObject,
  ): Promise<Contents.IModel> {
    url += URLExt.objectToQueryString(options);
    return requestAPI<Contents.IModel>(url, 'GET');
  }
}
