import {Pipeline, RepoInfo, PipelineInfo} from '@pachyderm/node-pachyderm';
import flattenDeep from 'lodash/flattenDeep';
import keyBy from 'lodash/keyBy';

import {UUID_WITHOUT_DASHES_REGEX} from '@dash-backend/constants/pachCore';
import flattenPipelineInput from '@dash-backend/lib/flattenPipelineInput';
import getJobsFromJobSet from '@dash-backend/lib/getJobsFromJobSet';
import hasRepoReadPermissions from '@dash-backend/lib/hasRepoReadPermissions';
import {QueryResolvers} from '@graphqlTypes';

import {
  jobInfosToGQLJobSet,
  pipelineInfoToGQLPipeline,
  repoInfoToGQLRepo,
} from './builders/pps';

interface SearchResolver {
  Query: {
    searchResults: QueryResolvers['searchResults'];
  };
}

const searchResolver: SearchResolver = {
  Query: {
    searchResults: async (
      _parent,
      {args: {query, limit, projectId, globalIdFilter}},
      {pachClient},
    ) => {
      if (query) {
        const lowercaseQuery = query.toLowerCase();
        let repos: RepoInfo.AsObject[];
        let pipelines: PipelineInfo.AsObject[];

        //Check if query is commit or job id
        if (UUID_WITHOUT_DASHES_REGEX.test(query)) {
          const jobSet = jobInfosToGQLJobSet(
            await pachClient.pps().inspectJobSet({id: query, projectId}),
            query,
          );

          if (jobSet.jobs.length === 0) {
            return {
              pipelines: [],
              repos: [],
              jobSet: null,
            };
          }

          return {
            pipelines: [],
            repos: [],
            jobSet,
          };
        }

        if (globalIdFilter) {
          const jobSet = await pachClient
            .pps()
            .inspectJobSet({id: globalIdFilter, projectId, details: false});

          const jobs = await getJobsFromJobSet({
            jobSet,
            projectId,
            pachClient,
          });

          const pipelineMap = jobs.reduce<Record<string, Pipeline.AsObject>>(
            (acc, job) => {
              if (job.job?.pipeline?.name) {
                return {...acc, [job.job?.pipeline?.name]: job.job?.pipeline};
              }
              return acc;
            },
            {},
          );

          const inputRepos = flattenDeep(
            jobs.map((job) =>
              job.details?.input ? flattenPipelineInput(job.details.input) : [],
            ),
          );

          repos = (await pachClient.pfs().listRepo()).filter(
            (repo) =>
              repo.repo?.name &&
              (pipelineMap[repo.repo?.name] ||
                inputRepos.includes(repo.repo?.name)),
          );

          pipelines = (await pachClient.pps().listPipeline()).filter(
            (pipeline) =>
              pipeline.pipeline?.name && pipelineMap[pipeline.pipeline?.name],
          );
        } else {
          [repos, pipelines] = await Promise.all([
            pachClient.pfs().listRepo(),
            pachClient.pps().listPipeline(),
          ]);
        }

        const filteredRepos = repos.filter(
          (r) =>
            r.repo?.name.toLowerCase().startsWith(lowercaseQuery) &&
            hasRepoReadPermissions(r.authInfo?.permissionsList),
        );

        const authorizedRepoMap = keyBy(
          filteredRepos,
          (r) => r.repo?.name || '',
        );

        const filteredPipelines = pipelines.filter(
          (p) =>
            authorizedRepoMap[p.pipeline?.name || ''] &&
            p.pipeline?.name.toLowerCase().startsWith(lowercaseQuery),
        );

        return {
          pipelines: filteredPipelines
            .slice(0, limit || pipelines.length)
            .map(pipelineInfoToGQLPipeline),
          repos: filteredRepos
            .slice(0, limit || repos.length)
            .map(repoInfoToGQLRepo),
          jobSet: null,
        };
      }

      return {
        pipelines: [],
        repos: [],
        jobSet: null,
      };
    },
  },
};

export default searchResolver;
