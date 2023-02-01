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
  const id = 'Solar-Panel-Data-Sorting_montage';
  const name = 'montage';
  const projectId = 'Solar-Panel-Data-Sorting';

  describe('pipeline', () => {
    it('should return a pipeline for a given id and projectId', async () => {
      const {data, errors = []} = await executeQuery<PipelineQuery>(
        GET_PIPELINE_QUERY,
        {
          args: {id: name, projectId},
        },
      );

      expect(errors).toHaveLength(0);
      expect(data?.pipeline).toEqual(
        expect.objectContaining({
          id: id,
          name: name,
          description: 'Not my favorite pipeline',
          state: 'PIPELINE_FAILURE',
          outputBranch: 'master',
          egress: true,
          s3OutputRepo: `s3//montage`,
        }),
      );
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
        args: {projectIds: [projectId]},
      });

      expect(data?.pipelines).toHaveLength(2);
      expect(data?.pipelines[0]?.id).toBe('Solar-Panel-Data-Sorting_montage');
      expect(data?.pipelines[1]?.id).toBe('Solar-Panel-Data-Sorting_edges');
    });

    it('should return pipeline list filtered by globalId', async () => {
      const {data} = await executeQuery<PipelinesQuery>(GET_PIPELINES_QUERY, {
        args: {
          projectIds: [projectId],
          jobSetId: '33b9af7d5d4343219bc8e02ff44cd55a',
        },
      });

      expect(data?.pipelines).toHaveLength(1);
      expect(data?.pipelines[0]?.id).toBe('Solar-Panel-Data-Sorting_montage');
    });
  });

  describe('createPipeline', () => {
    const projectId = 'Solar-Power-Data-Logger-Team-Collab';
    it('should create a pipeline', async () => {
      const {data: pipeline, errors: pipelineErrors = []} =
        await executeQuery<PipelineQuery>(GET_PIPELINE_QUERY, {
          args: {id: 'test', projectId},
        });

      expect(pipeline).toBeNull();
      expect(pipelineErrors[0].extensions.code).toBe('NOT_FOUND');

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
      expect(data?.createPipeline.id).toBe('_test');
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
      expect(errors[0].extensions.details).toBe(
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

      expect(pipeline?.pipeline.description).toBe('');

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
