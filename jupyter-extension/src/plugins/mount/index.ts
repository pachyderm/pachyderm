import {
  ILayoutRestorer,
  JupyterFrontEnd,
  JupyterFrontEndPlugin,
} from '@jupyterlab/application';
import {IDocumentManager} from '@jupyterlab/docmanager';
import {IFileBrowserFactory} from '@jupyterlab/filebrowser';
import {INotebookTracker} from '@jupyterlab/notebook';

import {MountPlugin} from './mount';
import {IMountPlugin} from './types';

const mount: JupyterFrontEndPlugin<IMountPlugin> = {
  id: 'jupyterlab-pachyderm:mount',
  autoStart: true,
  requires: [
    IDocumentManager,
    IFileBrowserFactory,
    ILayoutRestorer,
    INotebookTracker,
  ],
  activate: (
    app: JupyterFrontEnd,
    manager: IDocumentManager,
    factory: IFileBrowserFactory,
    restorer: ILayoutRestorer,
    tracker: INotebookTracker,
  ): IMountPlugin => {
    return new MountPlugin(app, manager, factory, restorer, tracker);
  },
};

export default mount;
