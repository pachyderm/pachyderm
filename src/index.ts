import {
  JupyterFrontEnd,
  JupyterFrontEndPlugin
} from '@jupyterlab/application';

import { requestAPI } from './handler';

/**
 * Initialization data for the jupyterlab-pachyderm extension.
 */
const mount: JupyterFrontEndPlugin<void> = {
  id: 'jupyterlab-pachyderm:mount',
  autoStart: true,
  activate: (app: JupyterFrontEnd) => {
    console.log('The Mount plugin is active!');

    requestAPI<any>('get_example')
      .then(data => {
        console.log(data);
      })
      .catch(reason => {
        console.error(
          `The jupyterlab_pachyderm server extension appears to be missing.\n${reason}`
        );
      });
  }
};

const hub: JupyterFrontEndPlugin<void> = {
  id: 'jupyterlab-pachyderm:hub',
  autoStart: true,
  activate: (app: JupyterFrontEnd) => {
    console.log('The Hub plugin is active!');
  }
};

/**
 * Export the plugins as default.
 */
const plugins: JupyterFrontEndPlugin<any>[] = [mount, hub];

export default plugins;
