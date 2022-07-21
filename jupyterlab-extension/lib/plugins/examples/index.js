import { ILauncher } from '@jupyterlab/launcher';
import { init } from './examples';
const help = {
    id: 'jupyterlab-pachyderm:examples',
    autoStart: true,
    requires: [ILauncher],
    activate: (app, launcher) => {
        return init(app, launcher);
    },
};
export default help;
