import {Status} from '@grpc/grpc-js/build/src/constants';
import {ApolloError} from 'apollo-server-express';

import client from '@dash-backend/grpc/client';
import {BYTES_IN_GIG} from '@dash-backend/lib/constants';
import {PipelineState, ProjectStatus, QueryResolvers} from '@graphqlTypes';

interface ProjectsResolver {
  Query: {
    projects: QueryResolvers['projects'];
    projectDetails: QueryResolvers['projectDetails'];
  };
}

const projectsResolver: ProjectsResolver = {
  Query: {
    projects: async (
      _field,
      _args,
      {pachdAddress = '', authToken = '', log},
    ) => {
      const pachClient = client({pachdAddress, authToken, log});

      try {
        const projects = await pachClient.projects().listProject();

        log.info({
          eventSource: 'project resolver',
          event: 'returning projects',
        });

        return projects.projectInfoList.map((project) => {
          return {
            name: project.name,
            description: project.description,
            createdAt: project.createdat?.seconds || 0,
            status: project.status,
            id: project.id,
          };
        });
      } catch (err) {
        if ((err as ApolloError).extensions.grpcCode === Status.UNIMPLEMENTED) {
          log.info({
            eventSource: 'project resolver',
            event: 'returning default project',
          });

          const [repos, pipelines] = await Promise.all([
            pachClient.pfs().listRepo(),
            pachClient.pps().listPipeline(),
          ]);

          repos.sort(
            (a, b) => (a.created?.seconds || 0) - (b.created?.seconds || 0),
          );
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

          return [
            {
              name: 'Default',
              description: 'Default Pachyderm project.',
              status,
              createdAt,
              id: 'default',
            },
          ];
        } else {
          return err;
        }
      }
    },
    projectDetails: async (
      _field,
      {args: {projectId}},
      {pachdAddress = '', authToken = '', log},
    ) => {
      log.info({
        eventSource: 'projectDetails resolver',
        event: 'returning project details',
      });

      const pachClient = client({pachdAddress, authToken, projectId, log});

      const [repos, pipelines, jobs] = await Promise.all([
        pachClient.pfs().listRepo(),
        pachClient.pps().listPipeline(),
        pachClient.pps().listJobs(),
      ]);

      const sizeGBytes = repos.reduce(
        (sum, r) => sum + r.sizeBytes / BYTES_IN_GIG,
        0,
      );

      return {
        sizeGBytes,
        repoCount: repos.length,
        pipelineCount: pipelines.length,
        jobs: jobs.map((job) => ({
          id: job.job?.id || '',
          state: job.state,
          createdAt: job.started?.seconds || 0,
        })),
      };
    },
  },
};

export default projectsResolver;
