import {CREATE_PROJECT_MUTATION} from '@dash-frontend/mutations/CreateProject';
import {DELETE_PROJECT_PSEUDO_FORCE_MUTATION} from '@dash-frontend/mutations/DeleteProject';
import {UPDATE_PROJECT_MUTATION} from '@dash-frontend/mutations/UpdateProject';
import {GET_PROJECT_DETAILS_QUERY} from '@dash-frontend/queries/GetProjectDetailsQuery';
import {GET_PROJECT_STATUS_QUERY} from '@dash-frontend/queries/GetProjectStatusQuery';
import {Status} from '@grpc/grpc-js/build/src/constants';

import {
  executeMutation,
  executeQuery,
  mockServer,
} from '@dash-backend/testHelpers';
import {
  CreateProjectMutation,
  DeleteProjectAndResourcesMutation,
  ProjectDetails,
  ProjectStatusQuery,
  UpdateProjectMutation,
} from '@graphqlTypes';

describe('Projects Resolver', () => {
  describe('projectStatus', () => {
    it('should return healthy for a project that has no pipelines', async () => {
      const {data} = await executeQuery<ProjectStatusQuery>(
        GET_PROJECT_STATUS_QUERY,
        {id: 'Empty-Project'},
      );

      const projects = data?.projectStatus;

      // At the time of making there are only 10 pipelines across two projects.
      expect(projects).toEqual(
        expect.objectContaining({
          __typename: 'Project',
          id: 'Empty-Project',
          status: 'HEALTHY',
        }),
      );
    });
  });

  describe('projectDetails', () => {
    it('should return project details', async () => {
      const {data} = await executeQuery<{
        projectDetails: ProjectDetails;
      }>(GET_PROJECT_DETAILS_QUERY, {
        args: {projectId: 'Solar-Panel-Data-Sorting'},
      });

      const projectDetails = data?.projectDetails;

      expect(projectDetails?.sizeDisplay).toBe('3 kB');
      expect(projectDetails?.repoCount).toBe(3);
      expect(projectDetails?.pipelineCount).toBe(2);
      expect(projectDetails?.jobSets).toHaveLength(4);
    });
  });

  describe('createProject', () => {
    it('should create a Project', async () => {
      expect(mockServer.getState().projects).not.toHaveProperty(
        'new-test-project',
      );

      const {errors} = await executeMutation<CreateProjectMutation>(
        CREATE_PROJECT_MUTATION,
        {
          args: {
            name: 'new-test-project',
            description: 'desc',
          },
        },
      );

      expect(errors).toBeUndefined();
      expect(mockServer.getState().projects).toHaveProperty('new-test-project');
      expect(
        mockServer.getState().projects['new-test-project'].getDescription(),
      ).toBe('desc');
    });
    it('should return an error if a project with that name already exists', async () => {
      expect(mockServer.getState().projects).toHaveProperty(
        'Data-Cleaning-Process',
      );

      const {errors} = await executeMutation<CreateProjectMutation>(
        CREATE_PROJECT_MUTATION,
        {
          args: {
            name: 'Data-Cleaning-Process',
            description: 'desc',
          },
        },
      );

      expect(errors).toHaveLength(1);
      expect(errors?.[0].extensions.grpcCode).toEqual(Status.ALREADY_EXISTS);
      expect(errors?.[0].extensions.details).toBe(
        'project Data-Cleaning-Process already exists.',
      );
    });

    it('should update a Project', async () => {
      expect(mockServer.getState().projects).toHaveProperty(
        'Data-Cleaning-Process',
      );
      expect(
        mockServer
          .getState()
          .projects['Data-Cleaning-Process'].getDescription(),
      ).not.toBe('desc');

      const {errors} = await executeMutation<UpdateProjectMutation>(
        UPDATE_PROJECT_MUTATION,
        {
          args: {
            name: 'Data-Cleaning-Process',
            description: 'desc',
          },
        },
      );

      expect(errors).toBeUndefined();
      expect(mockServer.getState().projects).toHaveProperty(
        'Data-Cleaning-Process',
      );
      expect(
        mockServer
          .getState()
          .projects['Data-Cleaning-Process'].getDescription(),
      ).toBe('desc');
    });
    it('should return an error if updating a project with that name that does not exist', async () => {
      expect(mockServer.getState().projects).not.toHaveProperty(
        'this-project-does-not-exist',
      );

      const {errors} = await executeMutation<UpdateProjectMutation>(
        UPDATE_PROJECT_MUTATION,
        {
          args: {
            name: 'this-project-does-not-exist',
            description: 'desc',
          },
        },
      );

      expect(errors).toHaveLength(1);
      // expect(errors?.[0].extensions.grpcCode).toEqual(Status.NOT_FOUND); // this is not working???
      expect(errors?.[0].extensions.code).toBe('NOT_FOUND');
      expect(errors?.[0].message).toBe(
        'project this-project-does-not-exist not found',
      ); // this should be extensions.details ???
    });
  });

  describe('deleteProjectAndResources', () => {
    it('should delete a Project and its repos and pipelines', async () => {
      expect(mockServer.getState().projects).toHaveProperty(
        'Data-Cleaning-Process',
      );
      expect(
        mockServer.getState().pipelines['Data-Cleaning-Process'],
      ).toHaveLength(8);
      expect(mockServer.getState().repos['Data-Cleaning-Process']).toHaveLength(
        14,
      );

      const {errors} = await executeMutation<DeleteProjectAndResourcesMutation>(
        DELETE_PROJECT_PSEUDO_FORCE_MUTATION,
        {
          args: {
            name: 'Data-Cleaning-Process',
          },
        },
      );

      expect(errors).toBeUndefined();
      expect(
        mockServer.getState().pipelines['Data-Cleaning-Process'],
      ).toBeUndefined();
      expect(
        mockServer.getState().repos['Data-Cleaning-Process'],
      ).toBeUndefined();
      expect(mockServer.getState().projects).not.toHaveProperty(
        'Data-Cleaning-Process',
      );
    });
  });
});
