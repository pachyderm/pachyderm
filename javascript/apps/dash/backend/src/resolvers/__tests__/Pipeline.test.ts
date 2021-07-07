/* eslint-disable @typescript-eslint/naming-convention */
import {executeQuery} from '@dash-backend/testHelpers';
import {GET_PIPELINE_QUERY} from '@dash-frontend/queries/GetPipelineQuery';
import {PipelineQuery} from '@graphqlTypes';

describe('Pipeline resolver', () => {
  describe('pipeline', () => {
    const id = 'montage';
    const projectId = '1';

    it('should return a pipeline for a given id and projectId', async () => {
      const {data, errors = []} = await executeQuery<PipelineQuery>(
        GET_PIPELINE_QUERY,
        {
          args: {id, projectId},
        },
      );

      expect(errors.length).toBe(0);
      expect(data?.pipeline.id).toBe(id);
      expect(data?.pipeline.name).toBe(id);
      expect(data?.pipeline.description).toBe('Not my favorite pipeline');
      expect(data?.pipeline.state).toBe('PIPELINE_FAILURE');
      expect(data?.pipeline.outputBranch).toBe('master');
      expect(data?.pipeline.egress).toBe(true);
      expect(data?.pipeline.s3OutputRepo).toBe(`s3//${id}`);
      expect(data?.pipeline.schedulingSpec).toStrictEqual({
        __typename: 'SchedulingSpec',
        nodeSelectorMap: [
          {__typename: 'NodeSelector', key: 'disktype', value: 'ssd'},
        ],
        priorityClassName: 'high-priority',
      });
    });

    it('should return a NOT_FOUND error if a pipeline could not be found', async () => {
      const {data, errors = []} = await executeQuery<PipelineQuery>(
        GET_PIPELINE_QUERY,
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
