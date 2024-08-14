import uniqBy from 'lodash/uniqBy';

import {JobInfo, JobInfoDetails} from '@dash-frontend/api/pps';
import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import {Vertex} from '@dash-frontend/lib/DAG/convertPipelinesAndReposToVertices';
import {NodeType} from '@dash-frontend/lib/types';

import {checkCronInputs} from '../checkCronInputs';
import {getUnixSecondsFromISOString} from '../dateTime';

import {
  generateVertexId,
  flattenPipelineInput,
  VertexIdentifierType,
  egressNodeName,
  jobEgressType,
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

    const projectName = job.job?.pipeline?.project?.name || '';
    const pipelineName = job.job?.pipeline?.name || '';

    const pipelineVertex: Vertex = {
      id: generateVertexId(projectName, pipelineName),
      project: projectName,
      name: pipelineName,
      parents: inputs,
      type: derivePipelineType(job.details),
      access: true,
      state: undefined,
      nodeState: undefined,
      jobState: job.state,
      jobNodeState: restJobStateToNodeState(job.state),
      hasCronInput: checkCronInputs(job.details?.input),
    };

    if (!egressNodeName(job.details?.egress))
      return [...acc, ...inputRepoVertices, pipelineVertex];

    // Egress nodes are a visual representation of a pipeline setting that is
    // unique to our DAG, so we want to append a identifier to the name to have it
    // differ from the pipeline name.

    const egressName = egressNodeName(job.details?.egress);

    const egressNode: Vertex = {
      id: generateVertexId(projectName, `${pipelineName}.egress`),
      project: projectName,
      name: egressName || '',
      type: NodeType.EGRESS,
      egressType: jobEgressType(job.details),
      access: true,
      parents: [
        {
          id: generateVertexId(projectName, pipelineName),
          project: projectName,
          name: pipelineName,
          type: VertexIdentifierType.DEFAULT,
        },
      ],
      state: undefined,
      jobState: undefined,
      nodeState: undefined,
      jobNodeState: undefined,
      createdAt: getUnixSecondsFromISOString(job?.created),
    };

    return [...acc, ...inputRepoVertices, pipelineVertex, egressNode];
  }, []);

  return uniqBy(vertices, (v) => v.id);
};
