import {QueryResolvers} from '@graphqlTypes';

interface VersionResolver {
  Query: Required<Pick<QueryResolvers, 'versionInfo'>>;
}

const versionResolver: VersionResolver = {
  Query: {
    versionInfo: async (_parent, _args, {pachClient}) => {
      const versionInfo = await pachClient.version().getVersion();
      return {
        pachdVersion: versionInfo,
        consoleVersion: process.env.REACT_APP_RELEASE_VERSION,
      };
    },
  },
};

export default versionResolver;
