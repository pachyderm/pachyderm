import {GET_PROJECT_QUERY} from '@dash-frontend/queries/GetProjectQuery';
import {GET_PROJECTS_QUERY} from '@dash-frontend/queries/GetProjectsQuery';
import {GET_PROJECT_STATUS_QUERY} from '@dash-frontend/queries/GetProjectStatusQuery';
import {Status} from '@grpc/grpc-js/build/src/constants';

import projects from '@dash-backend/mock/fixtures/projects';
import {grpcServer} from '@dash-backend/mock/networkMock';
import {
  PfsAPIService,
  PfsIAPIServer,
  Project as ProjectProto,
  ProjectInfo,
  Pipeline,
  PipelineInfo,
  PpsAPIService,
  PipelineState,
  PpsIAPIServer,
} from '@dash-backend/proto';
import {executeQuery} from '@dash-backend/testHelpers';
import {Project, ProjectStatusQuery} from '@graphqlTypes';

describe('project', () => {
  it('should return a project for a given id', async () => {
    const service: Pick<PfsIAPIServer, 'inspectProject'> = {
      inspectProject: (call, callback) => {
        const project = new ProjectInfo()
          .setProject(new ProjectProto().setName('Solar-Panel-Data-Sorting'))
          .setDescription(
            'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
          );
        callback(null, project);
      },
    };
    grpcServer.addService(PfsAPIService, service);

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
    const service: Pick<PfsIAPIServer, 'inspectProject'> = {
      inspectProject: (call, callback) => {
        callback({code: Status.NOT_FOUND, details: 'Project not found'});
      },
    };
    grpcServer.addService(PfsAPIService, service);

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
    const service: Pick<PfsIAPIServer, 'listProject'> = {
      listProject: (call) => {
        call.write(
          new ProjectInfo()
            .setProject(new ProjectProto().setName('project1'))
            .setDescription('Lorem1'),
        );
        call.write(
          new ProjectInfo()
            .setProject(new ProjectProto().setName('project2'))
            .setDescription('Lorem2'),
        );
        call.end();
      },
    };

    grpcServer.addService(PfsAPIService, service);

    const {data} = await executeQuery<{
      projects: Project[];
    }>(GET_PROJECTS_QUERY);

    const projects = data?.projects;

    expect(projects).toEqual(
      expect.arrayContaining([
        {
          __typename: 'Project',
          createdAt: null,
          description: 'Lorem1',
          id: 'project1',
          status: null,
        },
        {
          __typename: 'Project',
          createdAt: null,
          description: 'Lorem2',
          id: 'project2',
          status: null,
        },
      ]),
    );
  });
});

describe('projectStatus', () => {
  it('should return project status', async () => {
    const service: Pick<PpsIAPIServer, 'listPipeline'> = {
      listPipeline: (call) => {
        call.write(
          new PipelineInfo()
            .setPipeline(
              new Pipeline().setProject(
                new ProjectProto().setName('Solar-Panel-Data-Sorting'),
              ),
            )
            .setState(PipelineState.PIPELINE_FAILURE),
        );
        call.end();
      },
    };
    grpcServer.addService(PpsAPIService, service);

    const {data, errors} = await executeQuery<ProjectStatusQuery>(
      GET_PROJECT_STATUS_QUERY,
      {id: 'Solar-Panel-Data-Sorting'},
    );

    expect(errors).toBeUndefined();

    const projects = data?.projectStatus;

    expect(projects).toEqual(
      expect.objectContaining({
        __typename: 'Project',
        id: 'Solar-Panel-Data-Sorting',
        status: 'UNHEALTHY',
      }),
    );
  });
});
