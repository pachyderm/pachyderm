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

      expect(projects).toHaveLength(7);
      expect(projects?.[0]?.id).toBe('Solar-Panel-Data-Sorting');
      expect(projects?.[1]?.id).toBe('Data-Cleaning-Process');
      expect(projects?.[2]?.id).toBe('Solar-Power-Data-Logger-Team-Collab');
      expect(projects?.[3]?.id).toBe('Solar-Price-Prediction-Modal');
      expect(projects?.[4]?.id).toBe('Egress-Examples');
      expect(projects?.[5]?.id).toBe('Empty-Project');
      expect(projects?.[6]?.id).toBe('Trait-Discovery');
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
