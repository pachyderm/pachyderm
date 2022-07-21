import { requestAPI } from './handler';
import telemetry from './plugins/telemetry';
import help from './plugins/help';
import examples from './plugins/examples';
import '@pachyderm/components/dist/style.css';
/**
 * Initialization data for the jupyterlab-pachyderm extension.
 */
const mount = {
    id: 'jupyterlab-pachyderm:mount',
    autoStart: true,
    activate: (app) => {
        console.log('The Mount plugin is active!');
        requestAPI('/v1/repos')
            .then((data) => {
            console.log(data);
        })
            .catch((reason) => {
            console.error(`The jupyterlab_pachyderm server extension appears to be missing.\n${reason}`);
        });
    },
};
/**
 * Export the plugins as default.
 */
const plugins = [
    mount,
    telemetry,
    help,
    examples,
];
export default plugins;
