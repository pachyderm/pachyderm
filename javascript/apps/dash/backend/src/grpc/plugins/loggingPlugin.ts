import Logger from 'bunyan';

import baseLogger from '@dash-backend/lib/log';
import {GRPCPlugin} from '@dash-backend/lib/types';

import isServiceError from '../utils/isServiceError';

const loggingPlugin: (log?: Logger) => GRPCPlugin = (log = baseLogger) => ({
  onCall: ({requestName}) => {
    log.info(`${requestName} request started`);
  },
  onCompleted: ({requestName}) => {
    log.info(`${requestName} request completed`);
  },
  onError: ({requestName, error}) => {
    if (isServiceError(error)) {
      log.error({error: error.details}, `${requestName} request failed`);
    } else {
      log.error({error}, `${requestName} request failed`);
    }
  },
});

export default loggingPlugin;
