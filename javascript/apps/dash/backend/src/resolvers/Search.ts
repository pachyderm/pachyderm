import {ApolloError} from 'apollo-server-express';

import {UUID_WITHOUT_DASHES_REGEX} from '@dash-backend/constants/pachCore';
import {QueryResolvers} from '@graphqlTypes';

import {jobInfoToGQLJob, pipelineInfoToGQLPipeline} from './builders/pps';

interface SearchResolver {
  Query: {
    searchResults: QueryResolvers['searchResults'];
  };
}

const searchResolver: SearchResolver = {
  Query: {
    searchResults: async (_parent, {query}, {pachClient, log}) => {
      if (query) {
        //Check if query is commit or job id
        if (UUID_WITHOUT_DASHES_REGEX.test(query)) {
          try {
            const job = await pachClient.pps().inspectJob(query);
            return {
              pipelines: [],
              job: job ? jobInfoToGQLJob(job) : null,
            };
          } catch (e) {
            if ((e as ApolloError).extensions.code === 'NOT_FOUND') {
              return {
                pipelines: [],
                job: null,
              };
            } else {
              return e;
            }
          }
        }

        const lowercaseQuery = query.toLowerCase();
        const jq = `select((.pipeline.name | ascii_downcase | contains("${lowercaseQuery}")) or (.description // "" | ascii_downcase | contains("${lowercaseQuery}")))`;

        log.info(
          {eventSource: 'searchResolver', meta: {jq}},
          'querying pipelines with jq filter',
        );

        const pipelines = await pachClient.pps().listPipeline(jq);
        return {
          pipelines: pipelines
            ? pipelines.map((p) => pipelineInfoToGQLPipeline(p))
            : [],
          job: null,
        };
      }

      return {
        pipelines: [],
        job: null,
      };
    },
  },
};

export default searchResolver;
