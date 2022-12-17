import {toProtoDatumState} from '@dash-backend/lib/gqlEnumMappers';
import {NotFoundError} from '@dash-backend/lib/types';
import {QueryResolvers} from '@graphqlTypes';

import {datumInfoToGQLDatum} from './builders/pps';

const DEFAULT_LIMIT = 100;

interface DatumResolver {
  Query: {
    datum: QueryResolvers['datum'];
    datums: QueryResolvers['datums'];
    datumSearch: QueryResolvers['datumSearch'];
  };
}

const datumResolver: DatumResolver = {
  Query: {
    datum: async (_parent, {args: {id, jobId, pipelineId}}, {pachClient}) => {
      const datum = await pachClient
        .pps()
        .inspectDatum({id, jobId, pipelineName: pipelineId});
      return datumInfoToGQLDatum(datum.toObject(), jobId);
    },
    datums: async (
      _parent,
      {args: {jobId, pipelineId, limit, filter, cursor}},
      {pachClient},
    ) => {
      limit = limit || DEFAULT_LIMIT;

      const enumFilter = filter
        ? filter.map((state) => toProtoDatumState(state))
        : undefined;

      const datums = await pachClient.pps().listDatums({
        jobId,
        pipelineName: pipelineId,
        filter: enumFilter,
        number: limit + 1,
        cursor: cursor || undefined,
      });

      let nextCursor = undefined;

      //If datums.length is not greater than limit there are no pages left
      if (datums.length > limit) {
        datums.pop(); //remove the extra datum from the response
        nextCursor = datums[datums.length - 1].datum?.id;
      }

      return {
        items: datums.map((datumInfo) => datumInfoToGQLDatum(datumInfo, jobId)),
        cursor: nextCursor,
        hasNextPage: !!nextCursor,
      };
    },
    datumSearch: async (
      _parent,
      {args: {id, jobId, pipelineId}},
      {pachClient},
    ) => {
      //TODO: Update once we get regex
      if (id.length !== 64) {
        return null;
      }
      try {
        const datum = await pachClient
          .pps()
          .inspectDatum({id, jobId, pipelineName: pipelineId});
        return datumInfoToGQLDatum(datum.toObject(), jobId);
      } catch (e) {
        if (e instanceof NotFoundError) {
          return null;
        }
        throw e;
      }
    },
  },
};

export default datumResolver;
