import {
  ILabShell,
  ILayoutRestorer,
  JupyterFrontEnd,
  JupyterFrontEndPlugin,
} from '@jupyterlab/application';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {IFileBrowserFactory} from '@jupyterlab/filebrowser';

import {MountPlugin} from './mount';
import {IMountPlugin} from './types';

const mount: JupyterFrontEndPlugin<IMountPlugin> = {
  id: 'jupyterlab-pachyderm:mount',
  autoStart: true,
  requires: [IDocumentManager, IFileBrowserFactory, ILayoutRestorer, ILabShell],
  activate: (
    app: JupyterFrontEnd,
    manager: IDocumentManager,
    factory: IFileBrowserFactory,
    restorer: ILayoutRestorer,
    widgetTracker: ILabShell,
  ): IMountPlugin => {
    return new MountPlugin(app, manager, factory, restorer, widgetTracker);
  },
};

export default mount;
