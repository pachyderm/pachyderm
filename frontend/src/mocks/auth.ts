import {
  GetAccountQuery,
  mockGetAuthorizeQuery,
  mockGetAccountQuery,
} from '@graphqlTypes';
import {graphql} from 'msw';

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
