import {pachydermClient} from '@pachyderm/node-pachyderm';
import {v4 as uuid} from 'uuid';

import errorPlugin from '@dash-backend/grpc/plugins/errorPlugin';
import loggingPlugin from '@dash-backend/grpc/plugins/loggingPlugin';
import baseLogger from '@dash-backend/lib/log';
import {Account} from '@graphqlTypes';

import {getAccountFromIdToken} from './auth';
const {GRPC_SSL} = process.env;

type createContextProps = {
  idToken: string;
  authToken: string;
  projectId: string;
  host?: string;
};

const createContext = async ({
  idToken,
  authToken,
  projectId,
  host = '',
}: createContextProps) => {
  let account: Account | undefined;

  if (idToken) {
    account = await getAccountFromIdToken(idToken);
  }

  const pachdAddress = process.env.PACHD_ADDRESS;
  const log = baseLogger.child({
    pachdAddress,
    operationId: uuid(),
    account: {
      id: account?.id,
      email: account?.email,
    },
  });

  const grpcLogger = baseLogger.child({
    eventSource: 'grpc client',
    pachdAddress,
    projectId,
  });

  const pachClient = pachydermClient({
    authToken,
    pachdAddress,
    projectId,
    plugins: [loggingPlugin(grpcLogger), errorPlugin],
    ssl: GRPC_SSL === 'true',
  });

  return {
    account,
    authToken,
    host,
    log,
    pachClient,
    pachdAddress,
  };
};

export default createContext;
