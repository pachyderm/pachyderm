import {GetAuthorizeQuery} from '@graphqlTypes';
import {graphql} from 'msw';

export const mockEmptyGetAuthorize = () =>
  graphql.query<GetAuthorizeQuery>('getAuthorize', (_req, res, ctx) => {
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
