import {status} from '@grpc/grpc-js';

import projects from '@dash-backend/mock/fixtures/projects';
import {
  mockServer,
  createServiceError,
  executeQuery,
} from '@dash-backend/testHelpers';
import {GET_PROJECT_DETAILS_QUERY} from '@dash-frontend/queries/GetProjectDetailsQuery';
import {GET_PROJECT_QUERY} from '@dash-frontend/queries/GetProjectQuery';
import {GET_PROJECTS_QUERY} from '@dash-frontend/queries/GetProjectsQuery';
import {Project, ProjectDetails} from '@graphqlTypes';

import {DEFAULT_PROJECT_ID} from '../Projects';

describe('Projects Resolver', () => {
  describe('project', () => {
    it('should return a project for a given id', async () => {
      const {data, errors = []} = await executeQuery<{project: Project}>(
        GET_PROJECT_QUERY,
        {id: '1'},
      );

      expect(errors?.length).toBe(0);
      expect(data?.project.name).toBe(projects['1'].getName());
    });

    it('should return a NOT_FOUND error if a project cannot be found', async () => {
      const {data, errors = []} = await executeQuery<{project: Project}>(
        GET_PROJECT_QUERY,
        {id: 'bologna'},
      );

      expect(data).toBeFalsy();
      expect(errors?.length).toBe(1);
      expect(errors[0].extensions.code).toBe('NOT_FOUND');
    });

    it('should return the default project', async () => {
      const {data, errors = []} = await executeQuery<{project: Project}>(
        GET_PROJECT_QUERY,
        {id: DEFAULT_PROJECT_ID},
      );

      expect(errors?.length).toBe(0);
      expect(data?.project.name).toBe('Default');
    });
  });
  describe('projects', () => {
    it('should return the default project when the project service is unimplemented', async () => {
      const error = createServiceError({code: status.UNIMPLEMENTED});
      mockServer.setProjectsError(error);

      const {data} = await executeQuery<{
        projects: Project[];
      }>(GET_PROJECTS_QUERY);

      const projects = data?.projects;

      expect(projects?.length).toEqual(1);
      expect(projects?.[0]?.name).toEqual('Default');
      expect(projects?.[0]?.id).toEqual('default');
      expect(projects?.[0]?.description).toEqual('Default Pachyderm project.');
    });

    it('should return projects when the project service is implemented', async () => {
      const {data} = await executeQuery<{
        projects: Project[];
      }>(GET_PROJECTS_QUERY);

      const projects = data?.projects;

      expect(projects?.length).toEqual(7);
      expect(projects?.[0]?.name).toEqual('Solar Panel Data Sorting');
      expect(projects?.[1]?.name).toEqual('Data Cleaning Process');
      expect(projects?.[2]?.name).toEqual(
        'Solar Power Data Logger Team Collab',
      );
      expect(projects?.[3]?.name).toEqual('Solar Price Prediction Modal');
      expect(projects?.[4]?.name).toEqual('Solar Industry Analysis 2020');
      expect(projects?.[5]?.name).toEqual('Empty Project');
      expect(projects?.[6]?.name).toEqual('Trait Discovery');
    });
  });

  describe('projectDetails', () => {
    it('should return project details', async () => {
      const {data} = await executeQuery<{
        projectDetails: ProjectDetails;
      }>(GET_PROJECT_DETAILS_QUERY, {args: {projectId: '1'}});

      const projectDetails = data?.projectDetails;

      expect(projectDetails?.sizeDisplay).toBe('2.93 KB');
      expect(projectDetails?.repoCount).toBe(3);
      expect(projectDetails?.pipelineCount).toBe(2);
      expect(projectDetails?.jobSets).toHaveLength(2);
    });
  });
});
