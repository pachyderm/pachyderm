import {JupyterFrontEndPlugin} from '@jupyterlab/application';

import help from './plugins/help';
import examples from './plugins/examples';
import mount from './plugins/mount';

/**
 * Export the plugins as default.
 */
const plugins: JupyterFrontEndPlugin<any>[] = [
  mount,
  help,
  examples,
];

export default plugins;
