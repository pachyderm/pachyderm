import React from 'react';
import {ILayoutRestorer, JupyterFrontEnd} from '@jupyterlab/application';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {SplitPanel} from '@lumino/widgets';
import {ReactWidget, UseSignal} from '@jupyterlab/apputils';
import {IFileBrowserFactory} from '@jupyterlab/filebrowser';
import {mountLogoIcon} from '../../utils/icons';

import SortableList from './components/SortableList';
import {MountDrive} from './mountDrive';
import {PollRepos} from './pollRepos';

export type Branch = {
  branch: string;
  mount: Mount;
};

export type Mount = {
  name: string | null;
  state: string | null;
  mode: string | null;
};

export type Repo = {
  repo: string;
  branches: Branch[];
};

const createFileBrowser = (
  app: JupyterFrontEnd,
  manager: IDocumentManager,
  factory: IFileBrowserFactory,
) => {
  const drive = new MountDrive(app.docRegistry);
  manager.services.contents.addDrive(drive);
  const browser = factory.createFileBrowser('jupyterlab-pachyderm-browser', {
    driveName: drive.name,
    state: null,
    refreshInterval: 10000,
  });
  return browser;
};

export const init = async (
  app: JupyterFrontEnd,
  manager: IDocumentManager,
  factory: IFileBrowserFactory,
  restorer: ILayoutRestorer,
): Promise<void> => {
  const open = (path: string) => {
    app.commands.execute('filebrowser:open-path', {
      path: 'mount-browser:' + path,
    });
  };

  const mountBrowser = createFileBrowser(app, manager, factory);

  const poller = new PollRepos('PollRepos');

  const mountedList = ReactWidget.create(
    <UseSignal signal={poller.mountedSignal}>
      {(_, mounted) => (
        <div className="pachyderm-mount-base">
          <div className="pachyderm-mount-base-title">Mounted Repositories</div>
          <SortableList open={open} repos={mounted ? mounted : []} />
        </div>
      )}
    </UseSignal>,
  );
  mountedList.addClass('pachyderm-mount-react-wrapper');

  const unmountedList = ReactWidget.create(
    <UseSignal signal={poller.unmountedSignal}>
      {(_, unmounted) => (
        <div className="pachyderm-mount-base">
          <div className="pachyderm-mount-base-title">
            Unmounted Repositories
          </div>
          <SortableList open={open} repos={unmounted ? unmounted : []} />
        </div>
      )}
    </UseSignal>,
  );
  unmountedList.addClass('pachyderm-mount-react-wrapper');

  const panel = new SplitPanel();
  panel.orientation = 'vertical';
  panel.spacing = 0;
  panel.title.icon = mountLogoIcon;
  panel.title.caption = 'Pachyderm Mount';
  panel.id = 'pachyderm-mount';
  panel.addWidget(mountedList);
  panel.addWidget(unmountedList);
  panel.addWidget(mountBrowser);
  SplitPanel.setStretch(mountedList, 1);
  SplitPanel.setStretch(unmountedList, 1);
  SplitPanel.setStretch(mountBrowser, 1);

  window.addEventListener('resize', () => {
    panel.update();
  });

  restorer.add(panel, 'jupyterlab-pachyderm');
  app.shell.add(panel, 'left', {rank: 100});
};
