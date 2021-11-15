import {JupyterFrontEnd, JupyterFrontEndPlugin} from '@jupyterlab/application';

import {init} from './telemetry';

const telemetry: JupyterFrontEndPlugin<void> = {
  id: 'jupyterlab-pachyderm:telemetry',
  autoStart: true,
  activate: (app: JupyterFrontEnd) => {
    init(app);
  },
};

export default telemetry;
