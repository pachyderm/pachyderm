import errorPlugin from '@dash-backend/grpc/plugins/errorPlugin';
import loggingPlugin from '@dash-backend/grpc/plugins/loggingPlugin';
import baseLogger from '@dash-backend/lib/log';
import {pachydermClient} from '@dash-backend/proto';

import {generateConsoleTraceUuid} from './generateTrace';

const getPachClient = (requestId: string) => {
  const grpcLogger = baseLogger.child({
    requestId,
    eventSource: 'grpc client',
    pachdAddress: process.env.PACHD_ADDRESS,
  });

  grpcLogger.info('Creating pach client');

  const pachClient = pachydermClient({
    pachdAddress: process.env.PACHD_ADDRESS,
    plugins: [loggingPlugin(grpcLogger), errorPlugin],
    ssl: process.env.GRPC_SSL === 'true',
  });

  return pachClient;
};

const getPachClientAndAttachHeaders = ({
  requestId,
  authToken,
  projectId,
}: {
  requestId: string;
  authToken?: string;
  projectId?: string; // this can be removed
}) => {
  if (!requestId) requestId = generateConsoleTraceUuid();

  const pachClient = getPachClient(requestId);

  pachClient.attachCredentials('x-request-id', requestId);
  if (authToken) pachClient.attachCredentials('authn-token', authToken);
  if (projectId) pachClient.attachCredentials('project-id', projectId);

  return pachClient;
};

export default getPachClientAndAttachHeaders;
