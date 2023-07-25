import {mockGetAuthorizeQuery, mockGetAccountQuery} from '@graphqlTypes';

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
