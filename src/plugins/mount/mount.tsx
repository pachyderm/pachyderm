import React from 'react';
import {ILayoutRestorer, JupyterFrontEnd} from '@jupyterlab/application';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {SplitPanel} from '@lumino/widgets';
import {ReactWidget, UseSignal} from '@jupyterlab/apputils';
import {IFileBrowserFactory} from '@jupyterlab/filebrowser';

import {mountLogoIcon} from '../../utils/icons';
import SortableList from './components/SortableList';
import {PollRepos} from './pollRepos';
import createCustomFileBrowser from './customFileBrowser';
import {IMountPlugin, Repo} from './types';

export const MOUNT_BROWSER_NAME = 'mount-browser:';

export class MountPlugin implements IMountPlugin {
  private _app: JupyterFrontEnd<JupyterFrontEnd.IShell, 'desktop' | 'mobile'>;
  private _poller: PollRepos;
  private _panel: SplitPanel;

  constructor(
    app: JupyterFrontEnd,
    manager: IDocumentManager,
    factory: IFileBrowserFactory,
    restorer: ILayoutRestorer,
  ) {
    this._app = app;
    this._poller = new PollRepos('PollRepos');

    const mountedList = ReactWidget.create(
      <UseSignal signal={this._poller.mountedSignal}>
        {(_, mounted) => (
          <div className="pachyderm-mount-base">
            <div className="pachyderm-mount-base-title">
              Mounted Repositories
            </div>
            <SortableList
              open={this.open}
              repos={mounted ? mounted : this._poller.mounted}
            />
          </div>
        )}
      </UseSignal>,
    );
    mountedList.addClass('pachyderm-mount-react-wrapper');

    const unmountedList = ReactWidget.create(
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
    unmountedList.addClass('pachyderm-mount-react-wrapper');

    const mountBrowser = createCustomFileBrowser(app, manager, factory);

    this._panel = new SplitPanel();
    this._panel.orientation = 'vertical';
    this._panel.spacing = 0;
    this._panel.title.icon = mountLogoIcon;
    this._panel.title.caption = 'Pachyderm Mount';
    this._panel.id = 'pachyderm-mount';
    this._panel.addWidget(mountedList);
    this._panel.addWidget(unmountedList);
    this._panel.addWidget(mountBrowser);
    SplitPanel.setStretch(mountedList, 1);
    SplitPanel.setStretch(unmountedList, 1);
    SplitPanel.setStretch(mountBrowser, 1);

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
}
