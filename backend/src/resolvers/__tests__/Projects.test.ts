import {GET_PROJECT_DETAILS_QUERY} from '@dash-frontend/queries/GetProjectDetailsQuery';
import {GET_PROJECT_QUERY} from '@dash-frontend/queries/GetProjectQuery';
import {GET_PROJECTS_QUERY} from '@dash-frontend/queries/GetProjectsQuery';
import {status} from '@grpc/grpc-js';

import projects from '@dash-backend/mock/fixtures/projects';
import {
  mockServer,
  createServiceError,
  executeQuery,
} from '@dash-backend/testHelpers';
import {Project, ProjectDetails} from '@graphqlTypes';

import {DEFAULT_PROJECT_ID} from '../Projects';

describe('Projects Resolver', () => {
  describe('project', () => {
    it('should return a project for a given id', async () => {
      const {data, errors = []} = await executeQuery<{project: Project}>(
        GET_PROJECT_QUERY,
        {id: '1'},
      );

      expect(errors).toHaveLength(0);
      expect(data?.project.name).toBe(projects['1'].getName());
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

    it('should return the default project', async () => {
      const {data, errors = []} = await executeQuery<{project: Project}>(
        GET_PROJECT_QUERY,
        {id: DEFAULT_PROJECT_ID},
      );

      expect(errors).toHaveLength(0);
      expect(data?.project.name).toBe('Default');
    });
  });
  describe('projects', () => {
    it('should return the default project when the project service is unimplemented', async () => {
      const error = createServiceError({code: status.UNIMPLEMENTED});
      mockServer.setError(error);

      const {data} = await executeQuery<{
        projects: Project[];
      }>(GET_PROJECTS_QUERY);

      const projects = data?.projects;

      expect(projects).toHaveLength(1);
      expect(projects?.[0]?.name).toBe('Default');
      expect(projects?.[0]?.id).toBe('default');
      expect(projects?.[0]?.description).toBe('Default Pachyderm project.');
    });

    it('should return projects when the project service is implemented', async () => {
      const {data} = await executeQuery<{
        projects: Project[];
      }>(GET_PROJECTS_QUERY);

      const projects = data?.projects;

      expect(projects).toHaveLength(7);
      expect(projects?.[0]?.name).toBe('Solar Panel Data Sorting');
      expect(projects?.[1]?.name).toBe('Data Cleaning Process');
      expect(projects?.[2]?.name).toBe('Solar Power Data Logger Team Collab');
      expect(projects?.[3]?.name).toBe('Solar Price Prediction Modal');
      expect(projects?.[4]?.name).toBe('Egress Examples');
      expect(projects?.[5]?.name).toBe('Empty Project');
      expect(projects?.[6]?.name).toBe('Trait Discovery');
    });
  });

  describe('projectDetails', () => {
    it('should return project details', async () => {
      const {data} = await executeQuery<{
        projectDetails: ProjectDetails;
      }>(GET_PROJECT_DETAILS_QUERY, {args: {projectId: '1'}});

      const projectDetails = data?.projectDetails;

      expect(projectDetails?.sizeDisplay).toBe('3 kB');
      expect(projectDetails?.repoCount).toBe(3);
      expect(projectDetails?.pipelineCount).toBe(2);
      expect(projectDetails?.jobSets).toHaveLength(4);
    });
  });
});
