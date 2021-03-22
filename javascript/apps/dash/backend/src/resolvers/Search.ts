import client from '@dash-backend/grpc/client';
import {QueryResolvers} from '@graphqlTypes';

import {pipelineInfoToGQLPipeline} from './builders/pps';

interface SearchResolver {
  Query: {
    searchResults: QueryResolvers['searchResults'];
  };
}

const searchResolver: SearchResolver = {
  Query: {
    searchResults: async (
      _parent,
      {query},
      {pachdAddress = '', authToken = '', log},
    ) => {
      if (query) {
        const lowercaseQuery = query.toLowerCase();
        const jq = `select((.pipeline.name | ascii_downcase | contains("${lowercaseQuery}")) or (.description // "" | ascii_downcase | contains("${lowercaseQuery}")))`;

        log.info(
          {eventSource: 'searchResolver', meta: {jq}},
          'querying pipelines with jq filter',
        );

        const pachClient = client({pachdAddress, authToken, log});
        const pipelines = await pachClient.pps().listPipeline(jq);

        return {
          pipelines: pipelines
            ? pipelines.map((p) => pipelineInfoToGQLPipeline(p))
            : [],
        };
      }
      return {
        pipelines: [],
      };
    },
  },
};

export default searchResolver;
