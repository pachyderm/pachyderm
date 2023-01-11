import {CREATE_PIPELINE_MUTATION} from '@dash-frontend/mutations/CreatePipeline';
import {GET_PIPELINE_QUERY} from '@dash-frontend/queries/GetPipelineQuery';
import {GET_PIPELINES_QUERY} from '@dash-frontend/queries/GetPipelinesQuery';
import {Status} from '@grpc/grpc-js/build/src/constants';

import {executeMutation, executeQuery} from '@dash-backend/testHelpers';
import {
  CreatePipelineMutation,
  PipelineQuery,
  PipelinesQuery,
} from '@graphqlTypes';

describe('Pipeline resolver', () => {
  const id = 'montage';
  const projectId = '1';

  describe('pipeline', () => {
    it('should return a pipeline for a given id and projectId', async () => {
      const {data, errors = []} = await executeQuery<PipelineQuery>(
        GET_PIPELINE_QUERY,
        {
          args: {id, projectId},
        },
      );

      expect(errors).toHaveLength(0);
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

      expect(errors).toHaveLength(1);
      expect(data).toBeNull();
      expect(errors[0].extensions.code).toBe('NOT_FOUND');
    });
  });

  describe('pipelines', () => {
    it('should return pipeline list', async () => {
      const {data} = await executeQuery<PipelinesQuery>(GET_PIPELINES_QUERY, {
        args: {projectId},
      });

      expect(data?.pipelines).toHaveLength(2);
      expect(data?.pipelines[0]?.id).toBe('montage');
      expect(data?.pipelines[1]?.id).toBe('edges');
    });

    it('should return pipeline list filtered by globalId', async () => {
      const {data} = await executeQuery<PipelinesQuery>(GET_PIPELINES_QUERY, {
        args: {projectId, jobSetId: '33b9af7d5d4343219bc8e02ff44cd55a'},
      });

      expect(data?.pipelines).toHaveLength(1);
      expect(data?.pipelines[0]?.id).toBe('montage');
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
            transform: {
              image: 'alpine',
              cmdList: ['sh'],
            },
            pfs: {
              name: 'images',
              repo: {name: 'images'},
            },
            projectId,
          },
        },
      );
      expect(errors).toHaveLength(0);
      expect(data?.createPipeline.id).toBe('test');
      expect(data?.createPipeline.name).toBe('test');
    });

    it('should return an error if a pipeline with that name already exists', async () => {
      const {errors = []} = await executeMutation<CreatePipelineMutation>(
        CREATE_PIPELINE_MUTATION,
        {
          args: {
            name: 'processor',
            transform: {
              image: 'alpine',
              cmdList: ['sh'],
            },
            pfs: {
              name: 'images',
              repo: {name: 'images'},
            },
            projectId,
          },
        },
      );
      expect(errors).toHaveLength(1);
      expect(errors[0].extensions.grpcCode).toEqual(Status.ALREADY_EXISTS);
      expect(errors[0].extensions.details).toEqual(
        'pipeline processor already exists',
      );
    });
    it('should update a pipeline', async () => {
      const {errors: creationErrors = []} =
        await executeMutation<CreatePipelineMutation>(
          CREATE_PIPELINE_MUTATION,
          {
            args: {
              name: 'test',
              transform: {
                image: 'alpine',
                cmdList: ['sh'],
              },
              pfs: {
                name: 'images',
                repo: {name: 'images'},
              },
              projectId,
            },
          },
        );
      expect(creationErrors).toHaveLength(0);

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
            transform: {
              image: 'alpine',
              cmdList: ['sh'],
            },
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
      expect(errors).toHaveLength(0);
      expect(data?.createPipeline.name).toBe('processor');
      expect(data?.createPipeline.description).toBe('test');
    });
  });
});
