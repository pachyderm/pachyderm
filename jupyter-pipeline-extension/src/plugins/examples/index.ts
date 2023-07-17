import {JupyterFrontEnd, JupyterFrontEndPlugin} from '@jupyterlab/application';
import {ILauncher} from '@jupyterlab/launcher';

import {init} from './examples';

const help: JupyterFrontEndPlugin<void> = {
  id: 'jupyterlab-pachyderm:examples',
  autoStart: true,
  requires: [ILauncher],
  activate: (app: JupyterFrontEnd, launcher: ILauncher) => {
    return init(app, launcher);
  },
};

export default help;
