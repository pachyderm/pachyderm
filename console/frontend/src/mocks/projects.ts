import {rest} from 'msw';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  ProjectInfo,
  ListProjectRequest,
  InspectProjectRequest,
} from '@dash-frontend/api/pfs';

type ProjectsMap = {
  projects: ProjectInfo[];
};

export const ALL_PROJECTS: ProjectsMap = {
  projects: [
    {
      project: {name: 'ProjectA'},
      description: 'A description for project a',
      createdAt: '2020-09-13T12:26:10.000Z',
    },
    {
      project: {name: 'ProjectB'},
      description: 'A description for project b',
      createdAt: '2017-07-14T02:40:20.000Z',
    },
    {
      project: {name: 'ProjectC'},
      description: 'A description for project c',
      createdAt: '2014-05-13T16:53:40.000Z',
    },
  ],
};

export const mockProjects = () =>
  rest.post<ListProjectRequest, Empty, ProjectInfo[]>(
    '/api/pfs_v2.API/ListProject',
    async (_req, res, ctx) => {
      return res(ctx.json(ALL_PROJECTS.projects));
    },
  );

export const mockInspectProject = () =>
  rest.post<InspectProjectRequest, Empty, ProjectInfo>(
    '/api/pfs_v2.API/InspectProject',
    async (_req, res, ctx) => {
      return res(
        ctx.json({
          project: {name: 'default'},
          description: '',
          createdAt: '2017-07-14T02:40:20.000Z',
        }),
      );
    },
  );
