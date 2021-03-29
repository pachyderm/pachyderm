import {status} from '@grpc/grpc-js';

import accounts from '@dash-backend/mock/fixtures/accounts';
import {
  mockServer,
  executeOperation,
  createServiceError,
} from '@dash-backend/testHelpers';
import {Account, AuthConfig, Tokens} from '@graphqlTypes';

describe('Auth resolver', () => {
  describe('exchangeCode', () => {
    const operationName = 'exchangeCode';
    const variables = {code: 'xyz'};

    it('should exchange an auth code for a pach token', async () => {
      const {data} = await executeOperation<{
        exchangeCode: Tokens;
      }>(operationName, variables);

      expect(data?.exchangeCode.pachToken).toBeTruthy();
    });

    it('should return an error if there is a problem exchanging the id token', async () => {
      const error = createServiceError({code: status.UNAUTHENTICATED});

      mockServer.setAuthError(error);

      const {data, errors = []} = await executeOperation<{
        exchangeCode: Tokens;
      }>(operationName, variables);

      expect(data).toBeNull();
      expect(errors.length).toBe(1);
      expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
    });

    it('should return an error if there is an issue with the IDP', async () => {
      mockServer.setTokenError(true);

      const {data, errors = []} = await executeOperation<{
        exchangeCode: Tokens;
      }>(operationName, variables);

      expect(data).toBeNull();
      expect(errors.length).toBe(1);
      expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
    });
  });
  describe('account', () => {
    it("should return the user's account info", async () => {
      const {data} = await executeOperation<{account: Account}>('getAccount');

      expect(data?.account).toStrictEqual(accounts['1']);
    });

    it('should return an error if the user is not authenticated', async () => {
      const {data, errors = []} = await executeOperation<{account: Account}>(
        'getAccount',
        {},
        {'id-token': ''},
      );

      expect(data).toBeNull();
      expect(errors.length).toBe(1);
      expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
    });
  });
  describe('authConfig', () => {
    it('should return the OIDC providers auth url', async () => {
      const {data, errors = []} = await executeOperation<{
        authConfig: AuthConfig;
      }>('authConfig', {}, {'id-token': ''});

      expect(data?.authConfig.authUrl).toBe(
        `http://localhost:${mockServer.state.authPort}/auth`,
      );
      expect(data?.authConfig.clientId).toBe(process.env.OAUTH_CLIENT_ID);
      expect(data?.authConfig.pachdClientId).toBe(
        process.env.OAUTH_PACHD_CLIENT_ID,
      );
      expect(errors.length).toBe(0);
    });

    it('should return an error if the OIDC provider is misconfigured', async () => {
      mockServer.setAuthConfigurationError(true);

      const {data, errors = []} = await executeOperation<{authUrl: string}>(
        'authConfig',
        {},
        {'id-token': ''},
      );

      expect(data).toBe(null);
      expect(errors[0].extensions.code).toBe('INTERNAL_SERVER_ERROR');
    });
  });
});
