import {DATUM_QUERY} from '@dash-frontend/queries/GetDatumQuery';
import {DATUMS_QUERY} from '@dash-frontend/queries/GetDatumsQuery';

import {executeQuery} from '@dash-backend/testHelpers';
import {DatumQuery, DatumsQuery, DatumState} from '@graphqlTypes';

describe('Datum resolver', () => {
  describe('datum', () => {
    it('should return a datum for a given job & pipeline & id', async () => {
      const {data, errors = []} = await executeQuery<DatumQuery>(DATUM_QUERY, {
        args: {
          projectId: '1',
          id: 'a',
          jobId: '33b9af7d5d4343219bc8e02ff44cd55a',
          pipelineId: 'montage',
        },
      });

      expect(errors.length).toBe(0);
      expect(data?.datum.state).toBe(DatumState.SUCCESS);
    });
  });

  describe('datums', () => {
    it('should return datums for a given job & pipeline', async () => {
      const {data, errors = []} = await executeQuery<DatumsQuery>(
        DATUMS_QUERY,
        {
          args: {
            projectId: '1',
            jobId: '33b9af7d5d4343219bc8e02ff44cd55a',
            pipelineId: 'montage',
          },
        },
      );

      expect(errors.length).toBe(0);
      expect(data?.datums?.length).toBe(2);
    });
  });
});
