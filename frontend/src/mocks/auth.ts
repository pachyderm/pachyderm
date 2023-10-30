import {
  mockGetAuthorizeQuery,
  mockGetAccountQuery,
  mockGetRolesQuery,
  mockAuthConfigQuery,
  mockExchangeCodeMutation,
  Permission,
} from '@graphqlTypes';

export const mockEmptyGetAuthorize = () =>
  mockGetAuthorizeQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        getAuthorize: {
          satisfiedList: [],
          missingList: [],
          authorized: null,
          principal: '',
        },
      }),
    );
  });

export const mockFalseGetAuthorize = () =>
  mockGetAuthorizeQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        getAuthorize: {
          satisfiedList: [],
          missingList: [],
          authorized: false,
          principal: '',
        },
      }),
    );
  });

export const mockTrueGetAuthorize = (permissionsList: Permission[] = []) =>
  mockGetAuthorizeQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        getAuthorize: {
          satisfiedList: permissionsList,
          missingList: [],
          authorized: true,
          principal: '',
        },
      }),
    );
  });

export const mockGetAccountUnauth = () =>
  mockGetAccountQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        account: {
          email: '',
          id: 'unauthenticated',
          name: 'User',
        },
      }),
    );
  });

export const mockGetAccountAuth = () =>
  mockGetAccountQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        account: {
          email: 'email@user.com',
          id: 'TestUsername',
          name: 'User Test',
        },
      }),
    );
  });

export const mockEmptyGetRoles = () =>
  mockGetRolesQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        getRoles: {
          roleBindings: [
            {
              principal: '',
              roles: [],
            },
          ],
        },
      }),
    );
  });

export const mockAuthConfig = () =>
  mockAuthConfigQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        authConfig: {
          authEndpoint: '/dex/auth',
          clientId: 'console-test',
          pachdClientId: 'pachd',
        },
      }),
    );
  });

export const mockExchangeCode = () =>
  mockExchangeCodeMutation((_req, res, ctx) => {
    return res(
      ctx.data({
        exchangeCode: {
          pachToken: 'abc',
          idToken: '123',
        },
      }),
    );
  });
