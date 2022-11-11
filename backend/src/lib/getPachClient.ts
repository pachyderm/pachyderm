import memoize from 'lodash/memoize';

import errorPlugin from '@dash-backend/grpc/plugins/errorPlugin';
import loggingPlugin from '@dash-backend/grpc/plugins/loggingPlugin';
import baseLogger from '@dash-backend/lib/log';
import {pachydermClient} from '@dash-backend/proto';

const memo = memoize(
  (ssl: string, pachdAddress: string, projectId: string, authToken: string) => {
    const grpcLogger = baseLogger.child({
      eventSource: 'grpc client',
      pachdAddress,
    });

    grpcLogger.info('Creating pach client');

    const pachClient = pachydermClient({
      pachdAddress: pachdAddress,
      plugins: [loggingPlugin(grpcLogger), errorPlugin],
      ssl: ssl === 'true',
    });

    pachClient.attachCredentials({projectId, authToken});

    return pachClient;
  },
  (a, b, c, d) => a + b + c + d,
);

const getPachClient = (projectId: string, authToken: string) =>
  memo(
    process.env.GRPC_SSL || '',
    process.env.PACHD_ADDRESS || '',
    projectId,
    authToken,
  );

export default getPachClient;
