import {ApolloError} from 'apollo-server-express';

import {UUID_WITHOUT_DASHES_REGEX} from '@dash-backend/constants/pachCore';
import {QueryResolvers} from '@graphqlTypes';

import {
  jobInfoToGQLJob,
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
      {args: {query, limit}},
      {pachClient, log},
    ) => {
      if (query) {
        //Check if query is commit or job id
        if (UUID_WITHOUT_DASHES_REGEX.test(query)) {
          try {
            const job = await pachClient.pps().inspectJob(query);
            return {
              pipelines: [],
              repos: [],
              job: job ? jobInfoToGQLJob(job) : null,
            };
          } catch (e) {
            if ((e as ApolloError).extensions.code === 'NOT_FOUND') {
              return {
                pipelines: [],
                repos: [],
                job: null,
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
        const filteredRepos = repos.filter((r) =>
          r.repo?.name.toLowerCase().startsWith(lowercaseQuery),
        );

        return {
          pipelines: pipelines
            ? pipelines
                .slice(0, limit || pipelines.length)
                .map((p) => pipelineInfoToGQLPipeline(p))
            : [],
          repos: filteredRepos
            .slice(0, limit || repos.length)
            .map((r) => repoInfoToGQLRepo(r)),
          job: null,
        };
      }

      return {
        pipelines: [],
        repos: [],
        job: null,
      };
    },
  },
};

export default searchResolver;
