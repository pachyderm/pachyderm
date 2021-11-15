import {JupyterFrontEnd, JupyterFrontEndPlugin} from '@jupyterlab/application';

import {requestAPI} from './handler';
import telemetry from './plugins/telemetry';
import help from './plugins/help';

import '@pachyderm/components/dist/style.css';

/**
 * Initialization data for the jupyterlab-pachyderm extension.
 */
const mount: JupyterFrontEndPlugin<void> = {
  id: 'jupyterlab-pachyderm:mount',
  autoStart: true,
  activate: (app: JupyterFrontEnd) => {
    console.log('The Mount plugin is active!');

    requestAPI<any>('/v1/repos')
      .then((data) => {
        console.log(data);
      })
      .catch((reason) => {
        console.error(
          `The jupyterlab_pachyderm server extension appears to be missing.\n${reason}`,
        );
      });
  },
};

/**
 * Export the plugins as default.
 */
const plugins: JupyterFrontEndPlugin<any>[] = [mount, telemetry, help];

export default plugins;
