import errorPlugin from '@dash-backend/lib/errorPlugin';
import baseLogger from '@dash-backend/lib/log';
import loggingPlugin from '@dash-backend/lib/loggingPlugin';
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
    plugins: [loggingPlugin(grpcLogger), errorPlugin],
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
