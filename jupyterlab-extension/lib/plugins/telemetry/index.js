import { init } from './telemetry';
const telemetry = {
    id: 'jupyterlab-pachyderm:telemetry',
    autoStart: true,
    activate: (app) => {
        init(app);
    },
};
export default telemetry;
