import React from 'react';
import {ILayoutRestorer, JupyterFrontEnd} from '@jupyterlab/application';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {SplitPanel} from '@lumino/widgets';
import {ReactWidget, UseSignal} from '@jupyterlab/apputils';
import {FileBrowser, IFileBrowserFactory} from '@jupyterlab/filebrowser';
import {settingsIcon, spreadsheetIcon} from '@jupyterlab/ui-components';
import {Signal} from '@lumino/signaling';
import {JSONObject} from '@lumino/coreutils';

import {mountLogoIcon} from '../../utils/icons';
import {PollMounts} from './pollMounts';
import createCustomFileBrowser from './customFileBrowser';
import {
  AuthConfig,
  IMountPlugin,
  Repo,
  Mount,
  CurrentDatumResponse,
  PfsInput,
  ListMountsResponse,
  CrossInputSpec,
} from './types';
import Config from './components/Config/Config';
import Datum from './components/Datum/Datum';
import SortableList from './components/SortableList/SortableList';
import LoadingDots from '../../utils/components/LoadingDots/LoadingDots';
import FullPageError from './components/FullPageError/FullPageError';
import {requestAPI} from '../../handler';

export const MOUNT_BROWSER_NAME = 'mount-browser:';

export class MountPlugin implements IMountPlugin {
  private _app: JupyterFrontEnd<JupyterFrontEnd.IShell, 'desktop' | 'mobile'>;
  private _loader: ReactWidget;
  private _fullPageError: ReactWidget;
  private _config: ReactWidget;
  private _mountedList: ReactWidget;
  private _unmountedList: ReactWidget;
  private _datum: ReactWidget;
  private _mountBrowser: FileBrowser;
  private _poller: PollMounts;
  private _panel: SplitPanel;

  private _showConfig = false;
  private _showConfigSignal = new Signal<this, boolean>(this);
  private _readyPromise: Promise<void> = Promise.resolve();

  private _showDatum = false;
  private _keepMounted = false;
  private _currentDatumInfo: CurrentDatumResponse | undefined;
  private _showDatumSignal = new Signal<this, boolean>(this);
  private _repoViewInputSpec: CrossInputSpec | PfsInput = {};
  private _saveInputSpecSignal = new Signal<this, CrossInputSpec | PfsInput>(
    this,
  );

  constructor(
    app: JupyterFrontEnd,
    manager: IDocumentManager,
    factory: IFileBrowserFactory,
    restorer: ILayoutRestorer,
  ) {
    this._app = app;
    this._poller = new PollMounts('PollMounts');
    this._repoViewInputSpec = {};

    // This is used to detect if the config goes bad (pachd address changes)
    this._poller.configSignal.connect((_, config) => {
      if (config.cluster_status === 'INVALID' && !this._showConfig) {
        this.setShowConfig(true);
      }
    });

    // This is used to detect if the user becomes unauthenticated of there are errors on the server
    this._poller.statusSignal.connect((_, status) => {
      if (status.code === 500) {
        this.setShowFullPageError(true);
      }

      if (status.code === 401 && !this._showConfig) {
        this.setShowConfig(true);
      }
    });

    this._readyPromise = this.setup();

    this._config = ReactWidget.create(
      <UseSignal signal={this._showConfigSignal}>
        {(_, showConfig) => (
          <UseSignal signal={this._poller.configSignal}>
            {(_, authConfig) => (
              <UseSignal signal={this._poller.statusSignal}>
                {(_, status) => (
                  <Config
                    showConfig={showConfig ? showConfig : this._showConfig}
                    setShowConfig={this.setShowConfig}
                    reposStatus={
                      status ? status.code : this._poller.status.code
                    }
                    updateConfig={this.updateConfig}
                    authConfig={authConfig ? authConfig : this._poller.config}
                    refresh={this._poller.refresh}
                  />
                )}
              </UseSignal>
            )}
          </UseSignal>
        )}
      </UseSignal>,
    );
    this._config.addClass('pachyderm-mount-config-wrapper');

    this._mountedList = ReactWidget.create(
      <UseSignal signal={this._poller.mountedSignal}>
        {(_, mounted) => (
          <div className="pachyderm-mount-base">
            <div className="pachyderm-mount-config-container">
              <div className="pachyderm-mount-base-title">
                Mounted Repositories
              </div>
              <div style={{display: 'flex'}}>
                <button
                  className="pachyderm-button-link"
                  data-testid="Datum__mode"
                  onClick={() => this.setShowDatum(true)}
                  style={{marginRight: '0.25rem'}}
                >
                  Datum{' '}
                  <spreadsheetIcon.react
                    tag="span"
                    className="pachyderm-mount-icon-padding"
                  />
                </button>
                <button
                  className="pachyderm-button-link"
                  onClick={() => this.setShowConfig(true)}
                >
                  Config{' '}
                  <settingsIcon.react
                    tag="span"
                    className="pachyderm-mount-icon-padding"
                  />
                </button>
              </div>
            </div>
            <SortableList
              open={this.open}
              items={mounted ? mounted : this._poller.mounted}
              updateData={this._poller.updateData}
              mountedItems={[]}
            />
          </div>
        )}
      </UseSignal>,
    );
    this._mountedList.addClass('pachyderm-mount-react-wrapper');

    this._unmountedList = ReactWidget.create(
      <UseSignal signal={this._poller.unmountedSignal}>
        {(_, unmounted) => (
          <div className="pachyderm-mount-base">
            <div className="pachyderm-mount-base-title">
              Unmounted Repositories
            </div>
            <SortableList
              open={this.open}
              items={unmounted ? unmounted : this._poller.unmounted}
              updateData={this._poller.updateData}
              mountedItems={this._poller.mounted}
            />
          </div>
        )}
      </UseSignal>,
    );
    this._unmountedList.addClass('pachyderm-mount-react-wrapper');

    this._datum = ReactWidget.create(
      <UseSignal signal={this._showDatumSignal}>
        {(_, showDatum) => (
          <UseSignal signal={this._saveInputSpecSignal}>
            {(_, repoViewInputSpec) => (
              <Datum
                showDatum={showDatum ? showDatum : this._showDatum}
                setShowDatum={this.setShowDatum}
                keepMounted={this._keepMounted}
                setKeepMounted={this.setKeepMounted}
                open={this.open}
                pollRefresh={this._poller.refresh}
                currentDatumInfo={this._currentDatumInfo}
                repoViewInputSpec={
                  repoViewInputSpec
                    ? repoViewInputSpec
                    : this._repoViewInputSpec
                }
              />
            )}
          </UseSignal>
        )}
      </UseSignal>,
    );
    this._datum.addClass('pachyderm-mount-datum-wrapper');

    this._loader = ReactWidget.create(<LoadingDots />);
    this._loader.addClass('pachyderm-mount-react-wrapper');

    this._fullPageError = ReactWidget.create(
      <UseSignal signal={this._poller.statusSignal}>
        {(_, status) => (
          <FullPageError status={status ? status : this._poller.status} />
        )}
      </UseSignal>,
    );
    this._fullPageError.addClass('pachyderm-mount-react-wrapper');

    this._mountBrowser = createCustomFileBrowser(app, manager, factory);
    this._poller.mountedSignal.connect(this.verifyBrowserPath);
    this._poller.mountedSignal.connect(this.refresh);
    this._poller.unmountedSignal.connect(this.refresh);

    this._panel = new SplitPanel();
    this._panel.orientation = 'vertical';
    this._panel.spacing = 0;
    this._panel.title.icon = mountLogoIcon;
    this._panel.title.caption = 'Pachyderm Mount';
    this._panel.id = 'pachyderm-mount';
    this._panel.addWidget(this._mountedList);
    this._panel.addWidget(this._unmountedList);
    this._panel.addWidget(this._datum);
    this._panel.addWidget(this._mountBrowser);
    this._panel.setRelativeSizes([1, 1, 3, 3]);

    this._panel.addWidget(this._loader);
    this._panel.addWidget(this._config);
    this._panel.addWidget(this._fullPageError);

    //default view: hide all till ready
    this._config.setHidden(true);
    this._fullPageError.setHidden(true);
    this._mountedList.setHidden(true);
    this._unmountedList.setHidden(true);
    this._datum.setHidden(true);
    this._mountBrowser.setHidden(true);

    window.addEventListener('resize', () => {
      this._panel.update();
    });

    restorer.add(this._panel, 'jupyterlab-pachyderm');
    app.shell.add(this._panel, 'left', {rank: 100});
  }

  open = (path: string): void => {
    this._app.commands.execute('filebrowser:open-path', {
      path: MOUNT_BROWSER_NAME + path,
    });
  };

  refresh = async (_: PollMounts, _data: Mount[] | Repo[]): Promise<void> => {
    await this._mountBrowser.model.refresh();
  };

  // Change back to root directory if in a mount that no longer exists
  verifyBrowserPath = (_: PollMounts, mounted: Mount[]): void => {
    if (this._mountBrowser.model.path === this._mountBrowser.model.rootPath) {
      return;
    }

    const currentMountDir = this._mountBrowser.model.path
      .split(MOUNT_BROWSER_NAME)[1]
      .split('/')[0];
    for (let i = 0; i < mounted.length; i++) {
      if (currentMountDir === mounted[i].name) {
        return;
      }
    }
    this.open('');
  };

  setShowDatum = async (shouldShow: boolean): Promise<void> => {
    if (shouldShow) {
      this._datum.setHidden(false);
      this._mountedList.setHidden(true);
      this._unmountedList.setHidden(true);
      this.saveMountedReposList();
    } else {
      this._datum.setHidden(true);
      this._mountedList.setHidden(false);
      this._unmountedList.setHidden(false);
      await this.restoreMountedReposList();
    }
    this._mountBrowser.setHidden(false);
    this._config.setHidden(true);
    this._fullPageError.setHidden(true);
    this._showDatum = shouldShow;
    this._showDatumSignal.emit(shouldShow);
  };

  saveMountedReposList = (): void => {
    const mounted = this._poller.mounted;
    const pfsInputs: PfsInput[] = [];

    for (let i = 0; i < mounted.length; i++) {
      const pfsInput: PfsInput = {
        pfs: {
          ...(mounted[i].branch !== 'master' && {
            name: `${mounted[i].repo}_${mounted[i].branch}`,
          }),
          repo: mounted[i].repo,
          ...(mounted[i].branch !== 'master' && {branch: mounted[i].branch}),
          glob: '/',
        },
      };
      pfsInputs.push(pfsInput);
    }

    if (mounted.length === 0) {
      // No mounted repos to save
      this._repoViewInputSpec = {};
    } else if (mounted.length === 1) {
      // Single mounted repo to save
      this._repoViewInputSpec = pfsInputs[0];
    } else {
      // Multiple mounted repos to save; use cross
      this._repoViewInputSpec = {
        cross: pfsInputs,
      };
    }

    this._saveInputSpecSignal.emit(this._repoViewInputSpec);
  };

  restoreMountedReposList = async (): Promise<void> => {
    const mounts: JSONObject[] = [];
    // Restoring from cross of multiple mounts
    if (
      Object.prototype.hasOwnProperty.call(this._repoViewInputSpec, 'cross') &&
      Array.isArray(this._repoViewInputSpec.cross)
    ) {
      for (let i = 0; i < this._repoViewInputSpec.cross.length; i++) {
        const pfsInput = (this._repoViewInputSpec.cross[i] as PfsInput).pfs;
        mounts.push({
          name: pfsInput.name ? pfsInput.name : pfsInput.repo,
          repo: pfsInput.repo,
          branch: pfsInput.branch ? pfsInput.branch : 'master',
          mode: 'ro',
        });
      }
    }
    // Restoring from single mount
    else if (
      Object.prototype.hasOwnProperty.call(this._repoViewInputSpec, 'pfs')
    ) {
      const pfsInput = (this._repoViewInputSpec as PfsInput).pfs;
      mounts.push({
        name: pfsInput.name ? pfsInput.name : pfsInput.repo,
        repo: pfsInput.repo,
        branch: pfsInput.branch ? pfsInput.branch : 'master',
        mode: 'ro',
      });
    }

    if (mounts.length > 0) {
      const res = await requestAPI<ListMountsResponse>('_mount', 'PUT', {
        mounts: mounts,
      });
      this._poller.updateData(res);
    }
    this.open('');
  };

  setKeepMounted = (keep: boolean): void => {
    this._keepMounted = keep;
  };

  setShowConfig = (shouldShow: boolean): void => {
    if (shouldShow) {
      this._config.setHidden(false);
      this._mountedList.setHidden(true);
      this._unmountedList.setHidden(true);
      this._mountBrowser.setHidden(true);
    } else {
      this._config.setHidden(true);
      this._mountedList.setHidden(false);
      this._unmountedList.setHidden(false);
      this._mountBrowser.setHidden(false);
    }
    this._datum.setHidden(true);
    this._fullPageError.setHidden(true);
    this._showConfig = shouldShow;
    this._showConfigSignal.emit(shouldShow);
  };

  setShowFullPageError = (shouldShow: boolean): void => {
    if (shouldShow) {
      this._fullPageError.setHidden(false);
      this._config.setHidden(true);
      this._datum.setHidden(true);
      this._mountedList.setHidden(true);
      this._unmountedList.setHidden(true);
      this._mountBrowser.setHidden(true);
    } else {
      this._fullPageError.setHidden(true);
      this._config.setHidden(false);
      this._datum.setHidden(false);
      this._mountedList.setHidden(false);
      this._unmountedList.setHidden(false);
      this._mountBrowser.setHidden(false);
    }
  };

  updateConfig = (config: AuthConfig): void => {
    this._poller.config = config;
  };

  setup = async (): Promise<void> => {
    await this._poller.refresh();

    if (this._poller.status.code === 500) {
      this.setShowFullPageError(true);
    } else {
      this.setShowConfig(
        this._poller.config.cluster_status === 'INVALID' ||
          this._poller.status.code !== 200,
      );

      if (this._poller.status.code === 200) {
        try {
          const res = await requestAPI<CurrentDatumResponse>('datums', 'GET');
          if (res['num_datums'] > 0) {
            this._keepMounted = true;
            this._currentDatumInfo = res;
            await this.setShowDatum(true);
          }
        } catch (e) {
          console.log(e);
        }
      }
    }
    this._loader.setHidden(true);
  };

  get mountedRepos(): Mount[] {
    this._poller.poll.tick;
    return this._poller.mounted;
  }

  get unmountedRepos(): Repo[] {
    this._poller.poll.tick;
    return this._poller.unmounted;
  }

  get layout(): SplitPanel {
    return this._panel;
  }

  get ready(): Promise<void> {
    return this._readyPromise;
  }
}
