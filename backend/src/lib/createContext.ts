import {ApolloError} from 'apollo-server-errors';

import baseLogger from '@dash-backend/lib/log';
import {Account} from '@graphqlTypes';

import {getAccountFromIdToken} from './auth';
import {generateConsoleTraceUuid} from './generateTrace';
import getPachClientAndAttachHeaders from './getPachClient';

type createContextProps = {
  idToken: string;
  authToken: string;
  projectId: string;
  host?: string;
  requestId: string;
};

const createContext = async ({
  idToken,
  authToken,
  projectId,
  requestId,
  host = '',
}: createContextProps) => {
  let account: Account | undefined;
  if (!requestId) requestId = generateConsoleTraceUuid();

  const pachClient = getPachClientAndAttachHeaders({
    requestId,
    authToken,
    projectId,
  });

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
    requestId,
    pachdAddress,
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
