import Logger from 'bunyan';

import baseLogger from '@dash-backend/lib/log';
import {GRPCPlugin} from '@dash-backend/lib/types';

const loggingPlugin: (log?: Logger) => GRPCPlugin = (log = baseLogger) => ({
  onCall: ({requestName}) => {
    log.info(`${requestName} request started`);
  },
  onCompleted: ({requestName}) => {
    log.info(`${requestName} request completed`);
  },
  onError: ({requestName, error}) => {
    log.error({error: error.message}, `${requestName} request failed`);
  },
});

export default loggingPlugin;
