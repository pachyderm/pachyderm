import { IMainMenu } from '@jupyterlab/mainmenu';
import { init } from './help';
const help = {
    id: 'jupyterlab-pachyderm:help',
    autoStart: true,
    requires: [IMainMenu],
    activate: (app, mainMenu) => {
        init(app, mainMenu);
    },
};
export default help;
