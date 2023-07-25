import {
  mockGetAuthorizeQuery,
  mockGetAccountQuery,
  mockGetRolesQuery,
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
