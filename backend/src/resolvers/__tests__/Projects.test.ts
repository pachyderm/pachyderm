import {GET_PROJECT_DETAILS_QUERY} from '@dash-frontend/queries/GetProjectDetailsQuery';
import {GET_PROJECT_QUERY} from '@dash-frontend/queries/GetProjectQuery';
import {GET_PROJECTS_QUERY} from '@dash-frontend/queries/GetProjectsQuery';

import projects from '@dash-backend/mock/fixtures/projects';
import {executeQuery} from '@dash-backend/testHelpers';
import {Project, ProjectDetails} from '@graphqlTypes';

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

      expect(projects).toHaveLength(9);
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
});
