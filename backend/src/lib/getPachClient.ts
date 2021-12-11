import {pachydermClient} from '@pachyderm/node-pachyderm';
import memoize from 'lodash/memoize';

import errorPlugin from '@dash-backend/grpc/plugins/errorPlugin';
import loggingPlugin from '@dash-backend/grpc/plugins/loggingPlugin';
import baseLogger from '@dash-backend/lib/log';

const memo = memoize(
  (ssl: string, pachdAddress: string) => {
    const grpcLogger = baseLogger.child({
      eventSource: 'grpc client',
      pachdAddress,
    });

    grpcLogger.info('Creating pach client');

    return pachydermClient({
      pachdAddress: pachdAddress,
      plugins: [loggingPlugin(grpcLogger), errorPlugin],
      ssl: ssl === 'true',
    });
  },
  (a, b) => a + b,
);

const getPachClient = () =>
  memo(process.env.GRPC_SSL || '', process.env.PACHD_ADDRESS || '');

export default getPachClient;
