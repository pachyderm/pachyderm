/* eslint-disable @typescript-eslint/naming-convention */
import {CREATE_PIPELINE_MUTATION} from '@dash-frontend/mutations/CreatePipeline';
import {GET_PIPELINE_QUERY} from '@dash-frontend/queries/GetPipelineQuery';
import {Status} from '@grpc/grpc-js/build/src/constants';

import {executeMutation, executeQuery} from '@dash-backend/testHelpers';
import {CreatePipelineMutation, PipelineQuery} from '@graphqlTypes';

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
  describe('createPipeline', () => {
    const projectId = '3';
    it('should create a pipeline', async () => {
      const {data: pipeline, errors: pipelineErrors = []} =
        await executeQuery<PipelineQuery>(GET_PIPELINE_QUERY, {
          args: {id: 'test', projectId},
        });

      expect(pipeline).toBeNull();
      expect(pipelineErrors[0].extensions.code).toEqual('NOT_FOUND');

      const {data, errors = []} = await executeMutation<CreatePipelineMutation>(
        CREATE_PIPELINE_MUTATION,
        {
          args: {
            name: 'test',
            image: 'alpine',
            cmdList: ['sh'],
            pfs: {
              name: 'images',
              repo: {name: 'images'},
            },
            projectId,
          },
        },
      );
      expect(errors.length).toBe(0);
      expect(data?.createPipeline.id).toBe('test');
      expect(data?.createPipeline.name).toBe('test');
    });

    it('should return an error if a pipeline with that name already exists', async () => {
      const {errors = []} = await executeMutation<CreatePipelineMutation>(
        CREATE_PIPELINE_MUTATION,
        {
          args: {
            name: 'processor',
            image: 'alpine',
            cmdList: ['sh'],
            pfs: {
              name: 'images',
              repo: {name: 'images'},
            },
            projectId,
          },
        },
      );
      expect(errors.length).toBe(1);
      expect(errors[0].extensions.grpcCode).toEqual(Status.ALREADY_EXISTS);
      expect(errors[0].extensions.details).toEqual(
        'pipeline processor already exists',
      );
    });
    it('should update a pipeline', async () => {
      const {data: pipeline} = await executeQuery<PipelineQuery>(
        GET_PIPELINE_QUERY,
        {
          args: {id: 'test', projectId},
        },
      );

      expect(pipeline?.pipeline.description).toEqual('');

      const {data, errors = []} = await executeMutation<CreatePipelineMutation>(
        CREATE_PIPELINE_MUTATION,
        {
          args: {
            name: 'processor',
            image: 'alpine',
            cmdList: ['sh'],
            pfs: {
              name: 'images',
              repo: {name: 'images'},
            },
            description: 'test',
            projectId,
            update: true,
          },
        },
      );
      expect(errors.length).toBe(0);
      expect(data?.createPipeline.name).toBe('processor');
      expect(data?.createPipeline.description).toBe('test');
    });
  });
});
