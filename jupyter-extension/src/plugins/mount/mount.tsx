import React from 'react';
import {ILayoutRestorer, JupyterFrontEnd} from '@jupyterlab/application';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {SplitPanel} from '@lumino/widgets';
import {ReactWidget, UseSignal} from '@jupyterlab/apputils';
import {FileBrowser, IFileBrowserFactory} from '@jupyterlab/filebrowser';
import {settingsIcon} from '@jupyterlab/ui-components';
import {Signal} from '@lumino/signaling';

import {mountLogoIcon} from '../../utils/icons';
import {PollRepos} from './pollRepos';
import createCustomFileBrowser from './customFileBrowser';
import {AuthConfig, IMountPlugin, Repo} from './types';
import Config from './components/Config/Config';
import SortableList from './components/SortableList/SortableList';
import LoadingDots from '../../utils/components/LoadingDots/LoadingDots';
import FullPageError from './components/FullPageError/FullPageError';

export const MOUNT_BROWSER_NAME = 'mount-browser:';

export class MountPlugin implements IMountPlugin {
  private _app: JupyterFrontEnd<JupyterFrontEnd.IShell, 'desktop' | 'mobile'>;
  private _loader: ReactWidget;
  private _fullPageError: ReactWidget;
  private _config: ReactWidget;
  private _mountedList: ReactWidget;
  private _unmountedList: ReactWidget;
  private _mountBrowser: FileBrowser;
  private _poller: PollRepos;
  private _panel: SplitPanel;

  private _showConfig = false;
  private _showConfigSignal = new Signal<this, boolean>(this);
  private _readyPromise: Promise<void> = Promise.resolve();

  constructor(
    app: JupyterFrontEnd,
    manager: IDocumentManager,
    factory: IFileBrowserFactory,
    restorer: ILayoutRestorer,
  ) {
    this._app = app;
    this._poller = new PollRepos('PollRepos');

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
          <>
            <UseSignal signal={this._poller.configSignal}>
              {(_, authConfig) => (
                <>
                  <UseSignal signal={this._poller.statusSignal}>
                    {(_, status) => (
                      <>
                        <Config
                          showConfig={
                            showConfig ? showConfig : this._showConfig
                          }
                          setShowConfig={this.setShowConfig}
                          reposStatus={
                            status ? status.code : this._poller.status.code
                          }
                          updateConfig={this.updateConfig}
                          authConfig={
                            authConfig ? authConfig : this._poller.config
                          }
                          refresh={this._poller.refresh}
                        />
                      </>
                    )}
                  </UseSignal>
                </>
              )}
            </UseSignal>
          </>
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
            <SortableList
              open={this.open}
              repos={mounted ? mounted : this._poller.mounted}
              updateData={this._poller.updateData}
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
              repos={unmounted ? unmounted : this._poller.unmounted}
              updateData={this._poller.updateData}
            />
          </div>
        )}
      </UseSignal>,
    );
    this._unmountedList.addClass('pachyderm-mount-react-wrapper');

    this._loader = ReactWidget.create(
      <>
        <LoadingDots />
      </>,
    );
    this._loader.addClass('pachyderm-mount-react-wrapper');

    this._fullPageError = ReactWidget.create(
      <UseSignal signal={this._poller.statusSignal}>
        {(_, status) => (
          <>
            <FullPageError status={status ? status : this._poller.status} />
          </>
        )}
      </UseSignal>,
    );
    this._fullPageError.addClass('pachyderm-mount-react-wrapper');

    this._mountBrowser = createCustomFileBrowser(app, manager, factory);

    this._panel = new SplitPanel();
    this._panel.orientation = 'vertical';
    this._panel.spacing = 0;
    this._panel.title.icon = mountLogoIcon;
    this._panel.title.caption = 'Pachyderm Mount';
    this._panel.id = 'pachyderm-mount';
    this._panel.addWidget(this._mountedList);
    this._panel.addWidget(this._unmountedList);
    this._panel.addWidget(this._mountBrowser);
    SplitPanel.setStretch(this._mountedList, 1);
    SplitPanel.setStretch(this._unmountedList, 1);
    SplitPanel.setStretch(this._mountBrowser, 3);

    this._panel.addWidget(this._loader);
    this._panel.addWidget(this._config);
    this._panel.addWidget(this._fullPageError);

    //default view: hide all till ready
    this._config.setHidden(true);
    this._fullPageError.setHidden(true);
    this._mountedList.setHidden(true);
    this._unmountedList.setHidden(true);
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
    this._fullPageError.setHidden(true);
    this._showConfig = shouldShow;
    this._showConfigSignal.emit(shouldShow);
  };

  setShowFullPageError = (shouldShow: boolean): void => {
    if (shouldShow) {
      this._fullPageError.setHidden(false);
      this._config.setHidden(true);
      this._mountedList.setHidden(true);
      this._unmountedList.setHidden(true);
      this._mountBrowser.setHidden(true);
    } else {
      this._fullPageError.setHidden(true);
      this._config.setHidden(false);
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
    }
    this._loader.setHidden(true);
  };

  get mountedRepos(): Repo[] {
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
