import Logger from 'bunyan';

import formatBytes from '@dash-backend/lib/formatBytes';
import getSizeBytes from '@dash-backend/lib/getSizeBytes';
import {toGQLProjectStatus} from '@dash-backend/lib/gqlEnumMappers';
import {PachClient} from '@dash-backend/lib/types';
import {PipelineState, Project} from '@dash-backend/proto';
import {ProjectStatus, QueryResolvers} from '@graphqlTypes';

import {jobSetsToGQLJobSets} from './builders/pps';

interface ProjectsResolver {
  Query: {
    project: QueryResolvers['project'];
    projects: QueryResolvers['projects'];
    projectDetails: QueryResolvers['projectDetails'];
  };
}

export const DEFAULT_PROJECT_ID = 'default';

const rpcToGraphqlProject = (project: Project.AsObject) => {
  return {
    name: project.name,
    description: project.description,
    createdAt: project.createdat?.seconds || 0,
    status: toGQLProjectStatus(project.status),
    id: project.id,
  };
};

const getDefaultProject = async (client: PachClient, log: Logger) => {
  log.info({
    eventSource: 'project resolver',
    event: 'returning default project',
  });

  const [repos, pipelines] = await Promise.all([
    client.pfs().listRepo(),
    client.pps().listPipeline(),
  ]);

  repos.sort((a, b) => (a.created?.seconds || 0) - (b.created?.seconds || 0));
  const createdAt =
    repos.length && repos[0].created?.seconds
      ? repos[0].created.seconds
      : Math.floor(Date.now() / 1000);
  const status = pipelines.some(
    (pipeline) =>
      pipeline.state === PipelineState.PIPELINE_CRASHING ||
      pipeline.state === PipelineState.PIPELINE_FAILURE,
  )
    ? ProjectStatus.UNHEALTHY
    : ProjectStatus.HEALTHY;

  return {
    name: 'Default',
    description: 'Default Pachyderm project.',
    status,
    createdAt,
    id: DEFAULT_PROJECT_ID,
  };
};

const projectsResolver: ProjectsResolver = {
  Query: {
    project: async (_field, {id}, {pachClient, log}) => {
      // TODO: Remove this once projects are implemented
      if (id === DEFAULT_PROJECT_ID) {
        return await getDefaultProject(pachClient, log);
      }

      return rpcToGraphqlProject(
        await pachClient.projects().inspectProject(id),
      );
    },
    projects: async (_field, _args, {pachClient, log}) => {
      try {
        const projects = await pachClient.projects().listProject();

        log.info({
          eventSource: 'project resolver',
          event: 'returning projects',
        });

        return projects.projectInfoList.map(rpcToGraphqlProject);
      } catch (error) {
        // If _anything_ goes wrong fetching projects, fallback to the
        // console-generated default project. Originally, this was checking
        // for a specific UNIMPLEMENTED grpc error. However, pachd will throw
        // an UNKNOWN error for a service with an unimplemented auth function.
        // Since projects do not exist, projects do not implement auth.

        // TODO - we should not need this logic in the future, as projects
        // will always come from core pach
        log.info({
          eventSource: 'project resolver',
          event: 'problem fetching projects, falling back to default project',
          error,
        });

        return [await getDefaultProject(pachClient, log)];
      }
    },
    projectDetails: async (
      _field,
      {args: {jobSetsLimit}},
      {pachClient, log},
    ) => {
      log.info({
        eventSource: 'projectDetails resolver',
        event: 'returning project details',
      });

      const [repos, pipelines, jobSets] = await Promise.all([
        pachClient.pfs().listRepo(),
        pachClient.pps().listPipeline(),
        pachClient.pps().listJobSets({limit: jobSetsLimit, details: false}),
      ]);

      const totalSizeBytes = repos.reduce((sum, r) => sum + getSizeBytes(r), 0);

      return {
        sizeBytes: totalSizeBytes,
        sizeDisplay: formatBytes(totalSizeBytes),
        repoCount: repos.length,
        pipelineCount: pipelines.length,
        jobSets: jobSetsToGQLJobSets(jobSets),
      };
    },
  },
};

export default projectsResolver;
