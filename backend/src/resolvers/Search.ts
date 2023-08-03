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
      {pachClient},
    ) => {
      if (query) {
        const lowercaseQuery = query.toLowerCase();

        //Check if query is commit or job id
        if (UUID_WITHOUT_DASHES_REGEX.test(query)) {
          const jobSet = jobInfosToGQLJobSet(
            await pachClient.pps.inspectJobSet({id: query, projectId}),
            query,
          );

          if (jobSet.jobs.length === 0) {
            return {
              pipelines: [],
              repos: [],
              jobSet: null,
            };
          }

          return {
            pipelines: [],
            repos: [],
            jobSet,
          };
        }

        const [repos, pipelines] = await Promise.all([
          pachClient.pfs.listRepo({projectIds: [projectId]}),
          pachClient.pps.listPipeline({projectIds: [projectId]}),
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
          (p) =>
            authorizedRepoMap[p.pipeline?.name || ''] &&
            p.pipeline?.name.toLowerCase().startsWith(lowercaseQuery),
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
