import {QueryResolvers} from '@graphqlTypes';

import {datumInfoToGQLDatum} from './builders/pps';

const DEFAULT_OFFSET = 0;
const DEFAULT_LIMIT = 100;

interface DatumResolver {
  Query: {
    datum: QueryResolvers['datum'];
    datums: QueryResolvers['datums'];
  };
}

const datumResolver: DatumResolver = {
  Query: {
    datum: async (_parent, {args: {id, jobId, pipelineId}}, {pachClient}) => {
      const datum = await pachClient
        .pps()
        .inspectDatum({id, jobId, pipelineName: pipelineId});
      return datumInfoToGQLDatum(datum.toObject());
    },
    datums: async (
      _parent,
      {args: {jobId, pipelineId, limit, offset}},
      {pachClient},
    ) => {
      limit = limit || DEFAULT_LIMIT;
      offset = offset || DEFAULT_OFFSET;

      const datums = await pachClient
        .pps()
        .listDatums({jobId, pipelineName: pipelineId});
      return datums
        .slice(offset, offset + limit)
        .map((datumInfo) => datumInfoToGQLDatum(datumInfo));
    },
  },
};

export default datumResolver;
