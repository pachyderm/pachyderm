import {ModelDB} from '@jupyterlab/observables';
import {Contents, ServerConnection} from '@jupyterlab/services';
import {Signal, ISignal} from '@lumino/signaling';
import {DocumentRegistry} from '@jupyterlab/docregistry';
import {URLExt} from '@jupyterlab/coreutils';
import {requestAPI} from '../../handler';

export class MountDrive implements Contents.IDrive {
  public _registry: DocumentRegistry;
  private _fileChanged = new Signal<this, Contents.IChangedArgs>(this);
  private _serverSettings = ServerConnection.makeSettings();
  private _isDisposed = false;
  modelDBFactory?: ModelDB.IFactory | undefined;

  constructor(registry: DocumentRegistry) {
    this._registry = registry;
  }

  get name(): 'mount-browser' {
    return 'mount-browser';
  }
  get fileChanged(): ISignal<this, Contents.IChangedArgs> {
    return this._fileChanged;
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
    try {
      const response = await requestAPI<Contents.IModel>(
        URLExt.join('pfs', localPath),
        'GET',
      );
      return response;
    } catch (e) {
      console.log('/pfs not found');
      return {
        name: '',
        path: '',
        last_modified: '2022-04-26T16:28:48.015858Z',
        created: '2022-04-26T16:28:48.015858Z',
        content: [],
        format: 'json',
        mimetype: '',
        size: undefined,
        writable: true,
        type: 'directory',
      };
    }
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
