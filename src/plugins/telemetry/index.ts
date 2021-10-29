import {
  JupyterFrontEnd,
  JupyterFrontEndPlugin
} from '@jupyterlab/application';
import { INotebookTracker } from '@jupyterlab/notebook';
import { ITerminalTracker } from '@jupyterlab/terminal';

import { init } from './telemetry';

const telemetry: JupyterFrontEndPlugin<void> = {
  id: 'jupyterlab-pachyderm:telemetry',
  autoStart: true,
  activate: (
    app: JupyterFrontEnd,
    notebook: INotebookTracker,
    terminal: ITerminalTracker
  ) => {
    init(app, notebook, terminal);
  },
  requires: [INotebookTracker, ITerminalTracker]
};

export default telemetry;
