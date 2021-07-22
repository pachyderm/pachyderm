import {ApolloError} from 'apollo-server-express';
import keyBy from 'lodash/keyBy';

import {UUID_WITHOUT_DASHES_REGEX} from '@dash-backend/constants/pachCore';
import hasRepoReadPermissions from '@dash-backend/lib/hasRepoReadPermissions';
import {QueryResolvers} from '@graphqlTypes';

import {
  jobInfosToGQLJobSet,
  pipelineInfoToGQLPipeline,
  repoInfoToGQLRepo,
} from './builders/pps';

interface SearchResolver {
  Query: {
    searchResults: QueryResolvers['searchResults'];
  };
}

const searchResolver: SearchResolver = {
  Query: {
    searchResults: async (
      _parent,
      {args: {query, limit, projectId}},
      {pachClient, log},
    ) => {
      if (query) {
        //Check if query is commit or job id
        if (UUID_WITHOUT_DASHES_REGEX.test(query)) {
          try {
            const jobSet = jobInfosToGQLJobSet(
              await pachClient.pps().inspectJobSet({id: query, projectId}),
            );

            return {
              pipelines: [],
              repos: [],
              jobSet,
            };
          } catch (e) {
            if ((e as ApolloError).extensions.code === 'NOT_FOUND') {
              return {
                pipelines: [],
                repos: [],
                jobSet: null,
              };
            }
            return e;
          }
        }

        const lowercaseQuery = query.toLowerCase();
        const jq = `select(.pipeline.name | ascii_downcase | startswith("${lowercaseQuery}"))`;

        log.info(
          {eventSource: 'searchResolver', meta: {jq}},
          'querying pipelines with jq filter',
        );

        const [repos, pipelines] = await Promise.all([
          pachClient.pfs().listRepo(),
          pachClient.pps().listPipeline(jq),
        ]);

        const filteredRepos = repos.filter(
          (r) =>
            r.repo?.name.toLowerCase().startsWith(lowercaseQuery) &&
            hasRepoReadPermissions(r.authInfo?.permissionsList),
        );

        const authorizedRepoMap = keyBy(
          filteredRepos,
          (r) => r.repo?.name || '',
        );

        const filteredPipelines = pipelines.filter(
          (p) => authorizedRepoMap[p.pipeline?.name || ''],
        );

        return {
          pipelines: filteredPipelines
            .slice(0, limit || pipelines.length)
            .map(pipelineInfoToGQLPipeline),
          repos: filteredRepos
            .slice(0, limit || repos.length)
            .map(repoInfoToGQLRepo),
          jobSet: null,
        };
      }

      return {
        pipelines: [],
        repos: [],
        jobSet: null,
      };
    },
  },
};

export default searchResolver;
