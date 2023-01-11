import {DATUM_QUERY} from '@dash-frontend/queries/GetDatumQuery';
import {DATUM_SEARCH_QUERY} from '@dash-frontend/queries/GetDatumSearchQuery';
import {DATUMS_QUERY} from '@dash-frontend/queries/GetDatumsQuery';

import {executeQuery} from '@dash-backend/testHelpers';
import {
  Datum,
  DatumQuery,
  DatumSearchQuery,
  DatumsQuery,
  DatumState,
} from '@graphqlTypes';

describe('Datum resolver', () => {
  describe('datum', () => {
    it('should return a datum for a given job & pipeline & id', async () => {
      const {data, errors = []} = await executeQuery<DatumQuery>(DATUM_QUERY, {
        args: {
          projectId: '1',
          id: '0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
          jobId: '23b9af7d5d4343219bc8e02ff44cd55a',
          pipelineId: 'montage',
        },
      });

      expect(errors).toHaveLength(0);
      expect(data?.datum.state).toBe(DatumState.SUCCESS);
    });
  });

  describe('datumSearch', () => {
    it('should return a datum based on search criteria', async () => {
      const {data, errors = []} = await executeQuery<DatumSearchQuery>(
        DATUM_SEARCH_QUERY,
        {
          args: {
            projectId: '1',
            id: '0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
            jobId: '23b9af7d5d4343219bc8e02ff44cd55a',
            pipelineId: 'montage',
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.datumSearch?.state).toBe(DatumState.SUCCESS);
    });
    it('should return null if datum does not exist', async () => {
      const {data, errors = []} = await executeQuery<DatumSearchQuery>(
        DATUM_SEARCH_QUERY,
        {
          args: {
            projectId: '1',
            id: '0752b20131461a62943112579333667000000fff4a01406021603bbc98b4255d',
            jobId: '23b9af7d5d4343219bc8e02ff44cd55a',
            pipelineId: 'montage',
          },
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.datumSearch).toBe(null);
    });
  });

  describe('datums', () => {
    it('should return datums for a given job & pipeline', async () => {
      const {data, errors = []} = await executeQuery<DatumsQuery>(
        DATUMS_QUERY,
        {
          args: {
            projectId: '1',
            jobId: '23b9af7d5d4343219bc8e02ff44cd55a',
            pipelineId: 'montage',
          },
        },
      );
      const datums = data?.datums.items as Datum[];
      expect(errors).toHaveLength(0);
      expect(datums).toHaveLength(4);
    });

    it('should return filtered datums for a given job & pipeline', async () => {
      const {data, errors = []} = await executeQuery<DatumsQuery>(
        DATUMS_QUERY,
        {
          args: {
            projectId: '1',
            jobId: '23b9af7d5d4343219bc8e02ff44cd55a',
            pipelineId: 'montage',
            filter: [DatumState.STARTING],
          },
        },
      );
      expect(errors).toHaveLength(0);
      const datums = data?.datums.items as Datum[];
      expect(datums).toHaveLength(1);
      expect(datums[0].id).toBe(
        '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
      );
    });
  });
  describe('paging', () => {
    it('should return the first page if no cursor is specified', async () => {
      const {data, errors = []} = await executeQuery<DatumsQuery>(
        DATUMS_QUERY,
        {
          args: {
            projectId: '2',
            jobId: '23b9af7d5d4343219bc8e02ff4acd33a',
            pipelineId: 'likelihoods',
            limit: 10,
          },
        },
      );
      const datums = data?.datums.items as Datum[];
      console.log(datums);
      expect(errors).toHaveLength(0);
      expect(datums).toHaveLength(10);
      expect(data?.datums.cursor).toBe(
        '9a00000000000000000000000000000000000000000000000000000000000000',
      );
      expect(data?.datums.hasNextPage).toBe(true);
    });

    it('should return the next page if cursor is specified', async () => {
      const {data, errors = []} = await executeQuery<DatumsQuery>(
        DATUMS_QUERY,
        {
          args: {
            projectId: '2',
            jobId: '23b9af7d5d4343219bc8e02ff4acd33a',
            pipelineId: 'likelihoods',
            limit: 27,
            cursor:
              '63a0000000000000000000000000000000000000000000000000000000000000',
          },
        },
      );
      const datums = data?.datums.items as Datum[];
      expect(errors).toHaveLength(0);
      expect(datums).toHaveLength(27);
      expect(data?.datums.cursor).toBe(
        '90a0000000000000000000000000000000000000000000000000000000000000',
      );
      expect(data?.datums.hasNextPage).toBe(true);
    });

    it('should return no cursor if there are no more datums after the requested page', async () => {
      const {data, errors = []} = await executeQuery<DatumsQuery>(
        DATUMS_QUERY,
        {
          args: {
            projectId: '2',
            jobId: '23b9af7d5d4343219bc8e02ff4acd33a',
            pipelineId: 'likelihoods',
            limit: 10,
            cursor:
              '92a0000000000000000000000000000000000000000000000000000000000000',
          },
        },
      );
      const datums = data?.datums.items as Datum[];
      expect(errors).toHaveLength(0);
      expect(datums).toHaveLength(7);
      expect(data?.datums.cursor).toBe(null);
      expect(data?.datums.hasNextPage).toBe(false);
    });
  });
});
