import {executeOperation} from '@dash-backend/testHelpers';
import {PipelineQuery} from '@graphqlTypes';

describe('Pipeline resolver', () => {
  describe('pipeline', () => {
    const id = 'montage';
    const projectId = '1';

    it('should return a pipeline for a given id and projectId', async () => {
      const {data, errors = []} = await executeOperation<PipelineQuery>(
        'pipeline',
        {
          args: {id, projectId},
        },
      );

      expect(errors.length).toBe(0);
      expect(data?.pipeline.id).toBe(id);
      expect(data?.pipeline.name).toBe(id);
    });

    it('should return a NOT_FOUND error if a pipeline could not be found', async () => {
      const {data, errors = []} = await executeOperation<PipelineQuery>(
        'pipeline',
        {
          args: {id: 'bogus', projectId},
        },
      );

      expect(errors.length).toBe(1);
      expect(data).toBeNull();
      expect(errors[0].extensions.code).toBe('NOT_FOUND');
    });
  });
});
