import {CREATE_PROJECT_MUTATION} from '@dash-frontend/mutations/CreateProject';
import {DELETE_PROJECT_PSEUDO_FORCE_MUTATION} from '@dash-frontend/mutations/DeleteProject';
import {UPDATE_PROJECT_MUTATION} from '@dash-frontend/mutations/UpdateProject';
import {GET_PROJECT_DETAILS_QUERY} from '@dash-frontend/queries/GetProjectDetailsQuery';
import {GET_PROJECT_QUERY} from '@dash-frontend/queries/GetProjectQuery';
import {GET_PROJECTS_QUERY} from '@dash-frontend/queries/GetProjectsQuery';
import {Status} from '@grpc/grpc-js/build/src/constants';

import projects from '@dash-backend/mock/fixtures/projects';
import {
  executeMutation,
  executeQuery,
  mockServer,
} from '@dash-backend/testHelpers';
import {
  CreateProjectMutation,
  DeleteProjectAndResourcesMutation,
  Project,
  ProjectDetails,
  UpdateProjectMutation,
} from '@graphqlTypes';

describe('Projects Resolver', () => {
  describe('project', () => {
    it('should return a project for a given id', async () => {
      const {data, errors = []} = await executeQuery<{project: Project}>(
        GET_PROJECT_QUERY,
        {id: 'Solar-Panel-Data-Sorting'},
      );

      expect(errors).toHaveLength(0);
      expect(data?.project.id).toBe(
        projects['Solar-Panel-Data-Sorting'].getProject()?.getName(),
      );
    });

    it('should return a NOT_FOUND error if a project cannot be found', async () => {
      const {data, errors = []} = await executeQuery<{project: Project}>(
        GET_PROJECT_QUERY,
        {id: 'bologna'},
      );

      expect(data).toBeFalsy();
      expect(errors).toHaveLength(1);
      expect(errors[0].extensions.code).toBe('NOT_FOUND');
    });
  });
  describe('projects', () => {
    it('should return projects when the project service is implemented', async () => {
      const {data} = await executeQuery<{
        projects: Project[];
      }>(GET_PROJECTS_QUERY);

      const projects = data?.projects;

      expect(projects).toHaveLength(10);
      expect(projects?.[0]).toEqual(
        expect.objectContaining({
          __typename: 'Project',
          description:
            'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
          id: 'Solar-Panel-Data-Sorting',
          status: 'UNHEALTHY',
        }),
      );
      expect(projects?.[1]).toEqual(
        expect.objectContaining({
          __typename: 'Project',
          description:
            'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
          id: 'Data-Cleaning-Process',
          status: 'UNHEALTHY',
        }),
      );
      expect(projects?.[2]).toEqual(
        expect.objectContaining({
          __typename: 'Project',
          description:
            'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
          id: 'Solar-Power-Data-Logger-Team-Collab',
          status: 'HEALTHY',
        }),
      );
      expect(projects?.[3]).toEqual(
        expect.objectContaining({
          __typename: 'Project',
          description:
            'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
          id: 'Solar-Price-Prediction-Modal',
          status: 'UNHEALTHY',
        }),
      );
      expect(projects?.[4]).toEqual(
        expect.objectContaining({
          __typename: 'Project',
          description:
            'Multiple pipelines outputting to different forms of egress',
          id: 'Egress-Examples',
          status: 'UNHEALTHY',
        }),
      );
      expect(projects?.[5]).toEqual(
        expect.objectContaining({
          __typename: 'Project',
          description:
            'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
          id: 'Empty-Project',
          status: 'HEALTHY',
        }),
      );
      expect(projects?.[6]).toEqual(
        expect.objectContaining({
          __typename: 'Project',
          description:
            'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
          id: 'Trait-Discovery',
          status: 'HEALTHY',
        }),
      );
      expect(projects?.[7]).toEqual(
        expect.objectContaining({
          __typename: 'Project',
          description:
            'Contains two DAGs spanning across this and Multi-Project-Pipeline-B',
          id: 'Multi-Project-Pipeline-A',
          status: 'HEALTHY',
        }),
      );
      expect(projects?.[8]).toEqual(
        expect.objectContaining({
          __typename: 'Project',
          description:
            'Contains two DAGs spanning across this and Multi-Project-Pipeline-A',
          id: 'Multi-Project-Pipeline-B',
          status: 'HEALTHY',
        }),
      );
      expect(projects?.[9]).toEqual(
        expect.objectContaining({
          __typename: 'Project',
          description: 'Project for testing frontend load',
          id: 'Load-Project',
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
