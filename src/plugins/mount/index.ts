import {
  ILayoutRestorer,
  JupyterFrontEnd,
  JupyterFrontEndPlugin,
} from '@jupyterlab/application';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {IFileBrowserFactory} from '@jupyterlab/filebrowser';

import {init} from './mount';

const mount: JupyterFrontEndPlugin<void> = {
  id: 'jupyterlab-pachyderm:mount',
  autoStart: true,
  requires: [IDocumentManager, IFileBrowserFactory, ILayoutRestorer],
  activate: (
    app: JupyterFrontEnd,
    manager: IDocumentManager,
    factory: IFileBrowserFactory,
    restorer: ILayoutRestorer,
  ) => {
    init(app, manager, factory, restorer);
  },
};

export default mount;
