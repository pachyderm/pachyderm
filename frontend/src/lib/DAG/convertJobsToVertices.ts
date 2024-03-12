import uniqBy from 'lodash/uniqBy';

import {JobInfo, JobInfoDetails} from '@dash-frontend/api/pps';
import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import {Vertex} from '@dash-frontend/lib/DAG/convertPipelinesAndReposToVertices';
import {NodeType} from '@dash-frontend/lib/types';

import {checkCronInputs} from '../checkCronInputs';

import {
  generateVertexId,
  flattenPipelineInput,
  VertexIdentifierType,
} from './DAGhelpers';

const derivePipelineType = (details?: JobInfoDetails) => {
  if (details?.spout) {
    return details?.spout.service
      ? NodeType.SPOUT_SERVICE_PIPELINE
      : NodeType.SPOUT_PIPELINE;
  } else if (details?.service) {
    return NodeType.SERVICE_PIPELINE;
  }
  return NodeType.PIPELINE;
};

// Creates a vertex list of pachyderm repos and pipelines composed only of Global ID items
export const getProjectGlobalIdVertices = async (
  sortedJobs: JobInfo[],
  projectId: string,
) => {
  const vertices = sortedJobs.reduce<Vertex[]>((acc, job) => {
    const inputs = job.details?.input
      ? flattenPipelineInput(job.details.input)
      : [];

    const cronInputs: string[] = [];
    inputs.forEach((input) => {
      if (input.type === VertexIdentifierType.CRON) {
        cronInputs.push(input.id);
      }
    });

    const inputRepoVertices: Vertex[] = inputs.map(({project, name}) => {
      const id = generateVertexId(project, name);
      let type = NodeType.REPO;

      if (projectId !== project) {
        type = NodeType.CROSS_PROJECT_REPO;
      } else if (
        cronInputs.length !== 0 &&
        cronInputs.find((cron_id) => cron_id === id)
      ) {
        type = NodeType.CRON_REPO;
      }

      return {
        id,
        project,
        name,
        parents: [],
        type,
        state: undefined,
        jobState: undefined,
        nodeState: undefined,
        jobNodeState: undefined,
        access: true,
      };
    });

    const pipelineVertex: Vertex = {
      id: generateVertexId(
        job.job?.pipeline?.project?.name || '',
        job.job?.pipeline?.name || '',
      ),
      project: job.job?.pipeline?.project?.name || '',
      name: job.job?.pipeline?.name || '',
      parents: inputs,
      type: derivePipelineType(job.details),
      access: true,
      state: undefined,
      nodeState: undefined,
      jobState: job.state,
      jobNodeState: restJobStateToNodeState(job.state),
      hasCronInput: checkCronInputs(job.details?.input),
    };

    return [...acc, ...inputRepoVertices, pipelineVertex];
  }, []);

  return uniqBy(vertices, (v) => v.id);
};
