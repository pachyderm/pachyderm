import {ModelDB} from '@jupyterlab/observables';
import {Contents, ServerConnection} from '@jupyterlab/services';
import {PartialJSONObject} from '@lumino/coreutils';
import {Signal, ISignal} from '@lumino/signaling';
import {showErrorMessage} from '@jupyterlab/apputils';
import {DocumentRegistry} from '@jupyterlab/docregistry';
import {URLExt} from '@jupyterlab/coreutils';
import {requestAPI} from '../../handler';
import {MOUNT_BROWSER_PREFIX} from './mount';

// How many Content li are visible at a given time for the user to scroll.
const VISIBLE_CONTENT_LI_COUNT = 500;
// How many visible Content li that are padded on the top and bottom of the core set of visible Content li.
// This padding enables the user to scroll while the Contents index changes.
const VISIBLE_CONTENT_LI_PADDING = 200;
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
  // Previous scroll position of the Contents DOM node as of the previous user scroll event.
  private _previousScrollTop: number;
  // True if the FileBrowser contents scrolling event listener has been setup, false if not. Avoids setting up multiple
  // scroll event listeners.
  private _hasScrollEventListener: boolean;

  constructor(
    registry: DocumentRegistry,
    path: string,
    nameSuffix: string,
    id: string,
    _rerenderFileBrowser: () => Promise<void>,
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
    this._rerenderFileBrowser = _rerenderFileBrowser;
    this._previousFilter = null;
    this._index = 0;
    this._previousScrollTop = 0;
    this._hasScrollEventListener = false;
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
    const url = URLExt.join(this._path, localPath);

    // If we have cached content return that
    if (localPath === this._cache.key && localPath) {
      return this._getCachedContent();
    }

    // Make a no-content request to determine the type of content being requested.
    // If the content type is not a directory, then we want to preserve our cache
    //   and not paginate the results.
    let shallowResponse;
    try {
      shallowResponse = await this._get(url, {content: '0'});
    } catch (e) {
      showErrorMessage('Get Error', url + ' not found');
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
    const now = Date.now();
    this._cache = {
      key: localPath,
      now,
      contents: [],
      filteredContents: [],
      shallowResponse,
    };
    this._loading.emit(true);
    const getOptions = {
      ...options,
      number: PAGINATION_NUMBER,
      content,
    };
    const response = await this._get(url, getOptions);
    this._cache.contents = response.content;
    this.filterContents();
    this.resetContentsNode();
    this._loading.emit(false);
    this._fetchNextPage(response, now, url, getOptions);
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

    this.setupScrollingHandler();

    const {start} = this.getContentsStart();
    const {end} = this.getContentsEnd();
    return {
      ...this._cache.shallowResponse,
      content: this._cache.filteredContents.slice(start, end),
    };
  }

  private _fetchNextPage(
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
      this.filterContents();

      // Trigger a change directory event without a path change to force re-render the FileBrowser, then fetch the next page of results async.
      this._fetchNextPage(
        nextResponse,
        timeOfLastDirectoryChange,
        url,
        getOptions,
      );
    })();
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

  // Gets the visible cached Contents start for slicing. atMin is true if the _index can no longer be decremented.
  private getContentsStart(): {start: number; atMin: boolean} {
    let start = (this._index - 1) * VISIBLE_CONTENT_LI_PADDING;
    let atMin = false;
    if (start < 0) {
      start = 0;
      atMin = true;
    }
    return {start, atMin};
  }

  // Gets the visible cached Contents end for slicing. atMax is true if the _index can no longer be incremented.
  private getContentsEnd(): {end: number; atMax: boolean} {
    let end =
      VISIBLE_CONTENT_LI_COUNT + (this._index + 1) * VISIBLE_CONTENT_LI_PADDING;
    let atMax = false;
    if (end > this._cache.filteredContents.length) {
      end = this._cache.filteredContents.length;
      atMax = true;
    }
    return {end, atMax};
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
    let ignoreNextScroll = false;
    contentsNode.addEventListener('scroll', () => {
      const scrollTop = contentsNode.scrollTop;
      const scrollDiff = scrollTop - this._previousScrollTop;
      this._previousScrollTop = scrollTop;

      const scrollHeight = contentsNode.scrollHeight;
      const scrollNextHeight = Math.round(scrollHeight * 0.8);
      const scrollPrevHeight = Math.round(scrollHeight * 0.2);

      const contentsListItemNode: HTMLStyleElement = contentsNode
        .childNodes[0] as any;
      const contentsListItemHeight = contentsListItemNode.clientHeight;

      // Only the user scrolling to the upper or lower bound should trigger a change in index, not
      // scrolls initiated by this JS client.
      if (ignoreNextScroll) {
        return;
      }

      // When the user scrolls to the upper 20% of the Contents DOM node we decrement the index,
      // rerender the FileBrowser, and offset the scrollTop by how many Contents become change visibility
      // with each index change.
      const {atMin} = this.getContentsStart();
      if (scrollTop < scrollPrevHeight && scrollDiff < 0 && !atMin) {
        ignoreNextScroll = true;
        this._index -= 1;
        this._rerenderFileBrowser().then(() => {
          this._previousScrollTop +=
            VISIBLE_CONTENT_LI_PADDING * contentsListItemHeight;
          contentsNode.scrollTop = this._previousScrollTop;
          ignoreNextScroll = false;
        });
      }

      // When the user scrolls to the lower 80% of the Contents DOM node we increment the index,
      // rerender the FileBrowser, and offset the scrollTop by how many Contents become change visibility
      // with each index change.
      const {atMax} = this.getContentsEnd();
      if (scrollTop > scrollNextHeight && scrollDiff > 0 && !atMax) {
        ignoreNextScroll = true;
        this._index += 1;
        this._rerenderFileBrowser().then(() => {
          this._previousScrollTop -=
            VISIBLE_CONTENT_LI_PADDING * contentsListItemHeight;
          contentsNode.scrollTop = this._previousScrollTop;
          ignoreNextScroll = false;
        });
      }
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
