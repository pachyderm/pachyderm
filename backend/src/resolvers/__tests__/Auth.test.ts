import {EXCHANGE_CODE_MUTATION} from '@dash-frontend/mutations/ExchangeCode';
import {GET_ACCOUNT_QUERY} from '@dash-frontend/queries/GetAccountQuery';
import {GET_AUTH_CONFIG_QUERY} from '@dash-frontend/queries/GetAuthConfigQuery';
import {status} from '@grpc/grpc-js';

import accounts from '@dash-backend/mock/fixtures/accounts';
import {
  mockServer,
  createServiceError,
  executeMutation,
  executeQuery,
} from '@dash-backend/testHelpers';
import {
  Account,
  AuthConfig,
  GetPermissionsArgs,
  ResourceType,
  Tokens,
} from '@graphqlTypes';

import authResolver from '../Auth';

describe('Auth resolver', () => {
  describe('exchangeCode', () => {
    const variables = {code: 'xyz'};

    it('should exchange an auth code for a pach token', async () => {
      const {data} = await executeMutation<{
        exchangeCode: Tokens;
      }>(EXCHANGE_CODE_MUTATION, variables);

      expect(data?.exchangeCode.pachToken).toBeTruthy();
    });

    it('should return an error if there is a problem exchanging the id token', async () => {
      const error = createServiceError({code: status.UNAUTHENTICATED});

      mockServer.setError(error);

      const {data, errors = []} = await executeMutation<{
        exchangeCode: Tokens;
      }>(EXCHANGE_CODE_MUTATION, variables);

      expect(data).toBeNull();
      expect(errors).toHaveLength(1);
      expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
    });

    it('should return an error if there is an issue with the IDP', async () => {
      mockServer.setTokenError(true);

      const {data, errors = []} = await executeMutation<{
        exchangeCode: Tokens;
      }>(EXCHANGE_CODE_MUTATION, variables);

      expect(data).toBeNull();
      expect(errors).toHaveLength(1);
      expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
    });
  });
  describe('account', () => {
    it("should return the user's account info", async () => {
      const {data} = await executeQuery<{account: Account}>(GET_ACCOUNT_QUERY);

      expect(data?.account.id).toStrictEqual(accounts['1'].id);
    });

    it('should return an error if the user is not authenticated', async () => {
      const {data, errors = []} = await executeQuery<{account: Account}>(
        GET_ACCOUNT_QUERY,
        {},
        {'id-token': ''},
      );

      expect(data).toBeNull();
      expect(errors).toHaveLength(1);
      expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
    });
  });
  describe('authConfig', () => {
    it('should return the OIDC providers auth url', async () => {
      const {data, errors = []} = await executeQuery<{
        authConfig: AuthConfig;
      }>(GET_AUTH_CONFIG_QUERY, {}, {'id-token': ''});

      expect(data?.authConfig.authEndpoint).toBe(`/auth`);
      expect(data?.authConfig.clientId).toBe(process.env.OAUTH_CLIENT_ID);
      expect(data?.authConfig.pachdClientId).toBe(
        process.env.OAUTH_PACHD_CLIENT_ID,
      );
      expect(errors).toHaveLength(0);
    });

    it('should return an error if the OIDC provider is misconfigured', async () => {
      mockServer.setAuthConfigurationError(true);

      const {data, errors = []} = await executeQuery<{authUrl: string}>(
        GET_AUTH_CONFIG_QUERY,
        {},
        {'id-token': ''},
      );

      expect(data).toBeNull();
      expect(errors[0].extensions.code).toBe('UNAUTHENTICATED');
    });
  });
  describe('getPermissions - unit test', () => {
    it('should return rolesList', async () => {
      // TODO; Ideally we have a proper integration test here but permissions
      // and rbac are not implemented in the mock server. It would be preferred
      // if we could mock the pachClient functions. This doesn't work right now
      // though since the mock server uses the mock pachclient.
      //
      // // import {pachydermClient} from '@dash-backend/proto';
      // jest.mock('@dash-backend/proto', () => {
      //   return {
      //     __esModule: true,
      //     pachydermClient: jest.fn(() => {
      //       return {
      //         auth: () => {
      //           // eslint-disable-next-line no-unused-labels
      //           getPermisisons: () => {
      //             return Promise.resolve({rolesList: ['clusterAdmin']});
      //           };
      //         },
      //       };
      //     }),
      //   };
      // });
      //
      // const response = await testServer.executeOperation({
      //   query: GET_PERMISSIONS_QUERY,
      //   variables: {args: {resource: {type: 'REPO', name: 'default/wow'}}},
      // });

      const getPermissions = jest
        .fn()
        .mockResolvedValue({rolesList: ['clusterAdmin']});

      const pachClient = {
        auth: {getPermissions},
      };

      const args: GetPermissionsArgs = {
        resource: {name: 'tutorial/images', type: ResourceType.REPO},
      };

      // @ts-expect-error ts2349 -- ?. should fix it though but ts still isn't happy.
      const resp = await authResolver.Query.getPermissions?.(
        null,
        {args},
        {
          pachClient,
        },
      );

      expect(getPermissions).toHaveBeenCalledWith({
        resource: {
          name: 'tutorial/images',
          type: 'REPO',
        },
      });
      expect(resp).toStrictEqual({rolesList: ['clusterAdmin']});
    });
  });
});
