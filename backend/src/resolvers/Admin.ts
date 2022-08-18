import {QueryResolvers} from '@graphqlTypes';

interface AdminResolver {
  Query: {
    adminInfo: QueryResolvers['adminInfo'];
  };
}

const adminResolver: AdminResolver = {
  Query: {
    adminInfo: async (_parent, _args, {pachClient}) => {
      const clusterInfo = await pachClient.admin().inspectCluster();
      return {
        clusterId: clusterInfo.id,
      };
    },
  },
};

export default adminResolver;
