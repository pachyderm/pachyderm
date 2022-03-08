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
import {requestAPI} from '../../handler';

export const MOUNT_BROWSER_NAME = 'mount-browser:';

export class MountPlugin implements IMountPlugin {
  private _app: JupyterFrontEnd<JupyterFrontEnd.IShell, 'desktop' | 'mobile'>;
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

    this._readyPromise = this.setup();

    this._config = ReactWidget.create(
      <UseSignal signal={this._showConfigSignal}>
        {(_, showConfig) => (
          <>
            <UseSignal signal={this._poller.authenticatedSignal}>
              {(_, authenticated) => (
                <>
                  <Config
                    showConfig={showConfig ? showConfig : this._showConfig}
                    setShowConfig={this.setShowConfig}
                    authenticated={
                      authenticated ? authenticated : this._poller.authenticated
                    }
                  />
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
                className="pachyderm-button "
                onClick={() => this.setShowConfig(true)}
              >
                Config <settingsIcon.react tag="span" />
              </button>
            </div>
            <SortableList
              open={this.open}
              repos={mounted ? mounted : this._poller.mounted}
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
            />
          </div>
        )}
      </UseSignal>,
    );
    this._unmountedList.addClass('pachyderm-mount-react-wrapper');

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
    SplitPanel.setStretch(this._mountBrowser, 1);

    this._panel.addWidget(this._config);

    //default view
    this._config.setHidden(true);
    this._mountedList.setHidden(false);
    this._unmountedList.setHidden(false);
    this._mountBrowser.setHidden(false);

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
    this._showConfig = shouldShow;
    this._showConfigSignal.emit(shouldShow);
  };

  setup = async (): Promise<void> => {
    // Decide what screen to load on render
    try {
      const currentConfig = await requestAPI<AuthConfig>('config', 'GET');
      if (currentConfig.pachd_address) {
        await requestAPI<Repo[]>('repos', 'GET');
      } else {
        this.setShowConfig(true);
      }
    } catch (e) {
      this.setShowConfig(true);
    }
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
