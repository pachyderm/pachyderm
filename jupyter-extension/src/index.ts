import {JupyterFrontEndPlugin} from '@jupyterlab/application';

import telemetry from './plugins/telemetry';
import help from './plugins/help';
import examples from './plugins/examples';
import mount from './plugins/mount';

/**
 * Export the plugins as default.
 */
const plugins: JupyterFrontEndPlugin<any>[] = [
  mount,
  telemetry,
  help,
  examples,
];

export default plugins;
