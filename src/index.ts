import {
  JupyterFrontEnd,
  JupyterFrontEndPlugin
} from '@jupyterlab/application';

/**
 * Initialization data for the jupyterlab-pachyderm extension.
 */
const plugin: JupyterFrontEndPlugin<void> = {
  id: 'jupyterlab-pachyderm:plugin',
  autoStart: true,
  activate: (app: JupyterFrontEnd) => {
    console.log('JupyterLab extension jupyterlab-pachyderm is activated!');
  }
};

export default plugin;
