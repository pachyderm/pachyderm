import {ApolloError} from 'apollo-server-errors';
import {v4 as uuid} from 'uuid';

import baseLogger from '@dash-backend/lib/log';
import {Account} from '@graphqlTypes';

import {getAccountFromIdToken} from './auth';
import getPachClient from './getPachClient';

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

  const pachClient = getPachClient(projectId, authToken);

  if (idToken) {
    account = undefined;
    account = await getAccountFromIdToken(idToken);
  } else {
    try {
      await pachClient.auth().whoAmI();
    } catch (e) {
      const {code} = (e as ApolloError).extensions;
      if (code === 'UNIMPLEMENTED') {
        account = {
          id: 'unauthenticated',
          email: '',
          name: 'User',
        };
      } else if (code !== 'UNAUTHENTICATED') {
        throw e;
      }
    }
  }

  const pachdAddress = process.env.PACHD_ADDRESS;
  const log = baseLogger.child({
    pachdAddress,
    operationId: uuid(),
    account: {
      id: account?.id,
      email: account?.email,
    },
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
