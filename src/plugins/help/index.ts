import {JupyterFrontEnd, JupyterFrontEndPlugin} from '@jupyterlab/application';
import {IMainMenu} from '@jupyterlab/mainmenu';

import {init} from './help';

const help: JupyterFrontEndPlugin<void> = {
  id: 'jupyterlab-pachyderm:help',
  autoStart: true,
  requires: [IMainMenu],
  activate: (app: JupyterFrontEnd, mainMenu?: IMainMenu) => {
    init(app, mainMenu);
  },
};

export default help;
