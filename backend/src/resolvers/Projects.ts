import {ApolloError} from 'apollo-server-errors';

import {DEFAULT_JOBS_LIMIT} from '@dash-backend/constants/limits';
import formatBytes from '@dash-backend/lib/formatBytes';
import getSizeBytes from '@dash-backend/lib/getSizeBytes';
import {PipelineInfo, PipelineState} from '@dash-backend/proto';
import {ProjectInfo} from '@dash-backend/proto/proto/pfs/pfs_pb';
import {MutationResolvers, ProjectStatus, QueryResolvers} from '@graphqlTypes';

import {jobsMapToGQLJobSets} from './builders/pps';

interface ProjectsResolver {
  Query: {
    project: QueryResolvers['project'];
    projects: QueryResolvers['projects'];
    projectStatus: QueryResolvers['projectStatus'];
    projectDetails: QueryResolvers['projectDetails'];
  };
  Mutation: {
    createProject: MutationResolvers['createProject'];
    updateProject: MutationResolvers['updateProject'];
    deleteProjectAndResources: MutationResolvers['deleteProjectAndResources'];
  };
}

const rpcToGraphqlProject = (project: ProjectInfo.AsObject) => ({
  id: project.project?.name || '',
  description: project.description,
});

const getProjectHealth = (pipelines: PipelineInfo.AsObject[]) => {
  const status = pipelines.some(
    (pipeline) =>
      pipeline.state === PipelineState.PIPELINE_CRASHING ||
      pipeline.state === PipelineState.PIPELINE_FAILURE,
  )
    ? ProjectStatus.UNHEALTHY
    : ProjectStatus.HEALTHY;

  return status;
};

const projectsResolver: ProjectsResolver = {
  Query: {
    project: async (_field, {id}, {pachClient}) => {
      const project = await pachClient.pfs().inspectProject(id);
      return rpcToGraphqlProject(project);
    },
    projects: async (_field, _args, {pachClient}) => {
      const projects = await pachClient.pfs().listProject();
      return projects.map((project) => rpcToGraphqlProject(project));
    },
    projectStatus: async (_field, {id}, {pachClient}) => {
      const projectPipelines = await pachClient
        .pps()
        .listPipeline({projectIds: [id], details: false});

      return {id, status: getProjectHealth(projectPipelines)};
    },
    projectDetails: async (
      _field,
      {args: {jobSetsLimit, projectId}},
      {pachClient},
    ) => {
      const [repos, pipelines, jobsMap] = await Promise.all([
        pachClient.pfs().listRepo({projectIds: [projectId]}),
        pachClient
          .pps()
          .listPipeline({projectIds: [projectId], details: false}),
        pachClient.pps().listJobSetServerDerivedFromListJobs({
          projectId,
          limit: jobSetsLimit || DEFAULT_JOBS_LIMIT,
        }),
      ]);

      const totalSizeBytes = repos.reduce((sum, r) => sum + getSizeBytes(r), 0);

      return {
        sizeBytes: totalSizeBytes,
        sizeDisplay: formatBytes(totalSizeBytes),
        repoCount: repos.length,
        pipelineCount: pipelines.length,
        jobSets: jobsMapToGQLJobSets(jobsMap),
        status: getProjectHealth(pipelines),
      };
    },
  },
  Mutation: {
    createProject: async (
      _parent,
      {args: {name, description}},
      {pachClient},
    ) => {
      await pachClient.pfs().createProject({
        name,
        description: description || undefined,
      });
      return {
        ...rpcToGraphqlProject(await pachClient.pfs().inspectProject(name)),
        status: ProjectStatus.HEALTHY,
      };
    },
    updateProject: async (
      _parent,
      {args: {name, description}},
      {pachClient},
    ) => {
      await pachClient.pfs().createProject({
        name,
        description: description,
        update: true,
      });
      return rpcToGraphqlProject(await pachClient.pfs().inspectProject(name));
    },
    deleteProjectAndResources: async (
      _parent,
      {args: {name}},
      {pachClient},
    ) => {
      try {
        await pachClient.pps().deletePipelines({projectIds: [name]});
        await pachClient.pfs().deleteRepos({projectIds: [name]});
        await pachClient.pfs().deleteProject({projectId: name});
        return true;
      } catch (e) {
        throw new ApolloError(`Failed to delete project '${name}'`);
      }
    },
  },
};

export default projectsResolver;
