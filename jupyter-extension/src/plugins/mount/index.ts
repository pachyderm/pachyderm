import {
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
  requires: [IDocumentManager, IFileBrowserFactory, ILayoutRestorer],
  activate: (
    app: JupyterFrontEnd,
    manager: IDocumentManager,
    factory: IFileBrowserFactory,
    restorer: ILayoutRestorer,
  ): IMountPlugin => {
    return new MountPlugin(app, manager, factory, restorer);
  },
};

export default mount;
