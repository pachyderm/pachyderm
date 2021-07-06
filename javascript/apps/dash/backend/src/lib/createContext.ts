import {v4 as uuid} from 'uuid';

import baseLogger from '@dash-backend/lib/log';
import {Account} from '@graphqlTypes';

import client from '../grpc/client';

import {getAccountFromIdToken} from './auth';

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
  const pachClient = client({
    authToken,
    log,
    pachdAddress,
    projectId,
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
