import {QueryResolvers} from '@graphqlTypes';

import {enterpriseInfoToGQLInfo} from './builders/enterprise';

interface EnterpriseResolver {
  Query: {
    enterpriseInfo: QueryResolvers['enterpriseInfo'];
  };
}

const enterpriseResolver: EnterpriseResolver = {
  Query: {
    enterpriseInfo: async (_parent, _args, {pachClient}) => {
      return enterpriseInfoToGQLInfo(await pachClient.enterprise().getState());
    },
  },
};

export default enterpriseResolver;
