import {status} from '@grpc/grpc-js';

import projects from '@dash-backend/mock/fixtures/projects';
import {
  mockServer,
  executeOperation,
  createServiceError,
} from '@dash-backend/testHelpers';
import {Project, ProjectDetails} from '@graphqlTypes';

import {DEFAULT_PROJECT_ID} from '../Projects';

describe('Projects Resolver', () => {
  describe('project', () => {
    const operationName = 'project';

    it('should return a project for a given id', async () => {
      const {data, errors = []} = await executeOperation<{project: Project}>(
        operationName,
        {id: '1'},
      );

      expect(errors?.length).toBe(0);
      expect(data?.project.name).toBe(projects['1'].getName());
    });

    it('should return a NOT_FOUND error if a project cannot be found', async () => {
      const {data, errors = []} = await executeOperation<{project: Project}>(
        operationName,
        {id: 'bologna'},
      );

      expect(data).toBeFalsy();
      expect(errors?.length).toBe(1);
      expect(errors[0].extensions.code).toBe('NOT_FOUND');
    });

    it('should return the default project', async () => {
      const {data, errors = []} = await executeOperation<{project: Project}>(
        operationName,
        {id: DEFAULT_PROJECT_ID},
      );

      expect(errors?.length).toBe(0);
      expect(data?.project.name).toBe('Default');
    });
  });
  describe('projects', () => {
    const operationName = 'projects';

    it('should return the default project when the project service is unimplemented', async () => {
      const error = createServiceError({code: status.UNIMPLEMENTED});
      mockServer.setProjectsError(error);

      const {data} = await executeOperation<{
        projects: Project[];
      }>(operationName);

      const projects = data?.projects;

      expect(projects?.length).toEqual(1);
      expect(projects?.[0]?.name).toEqual('Default');
      expect(projects?.[0]?.id).toEqual('default');
      expect(projects?.[0]?.description).toEqual('Default Pachyderm project.');
    });

    it('should return projects when the project service is implemented', async () => {
      const {data} = await executeOperation<{
        projects: Project[];
      }>(operationName);

      const projects = data?.projects;

      expect(projects?.length).toEqual(5);
      expect(projects?.[0]?.name).toEqual('Solar Panel Data Sorting');
      expect(projects?.[1]?.name).toEqual('Data Cleaning Process');
      expect(projects?.[2]?.name).toEqual(
        'Solar Power Data Logger Team Collab',
      );
      expect(projects?.[3]?.name).toEqual('Solar Price Prediction Modal');
      expect(projects?.[4]?.name).toEqual('Solar Industry Analysis 2020');
    });
  });

  describe('projectDetails', () => {
    const operationName = 'projectDetails';
    it('should return project details', async () => {
      const {data} = await executeOperation<{
        projectDetails: ProjectDetails;
      }>(operationName, {args: {projectId: '1'}});

      const projectDetails = data?.projectDetails;

      expect(projectDetails?.sizeDisplay).toBe('2.93 KB');
      expect(projectDetails?.repoCount).toBe(3);
      expect(projectDetails?.pipelineCount).toBe(2);
      expect(projectDetails?.jobs).toHaveLength(2);
    });
  });
});
