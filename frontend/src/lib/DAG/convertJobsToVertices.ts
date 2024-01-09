import uniqBy from 'lodash/uniqBy';

import {JobInfo} from '@dash-frontend/api/pps';
import {restJobStateToNodeState} from '@dash-frontend/api/utils/nodeStateMappers';
import {Vertex} from '@dash-frontend/lib/DAG/convertPipelinesAndReposToVertices';
import {NodeType} from '@dash-frontend/lib/types';

import {generateVertexId, flattenPipelineInput} from './DAGhelpers';

// Creates a vertex list of pachyderm repos and pipelines composed only of Global ID items
export const getProjectGlobalIdVertices = async (
  sortedJobs: JobInfo[],
  projectId: string,
) => {
  const vertices = sortedJobs.reduce<Vertex[]>((acc, job) => {
    const inputs = job.details?.input
      ? flattenPipelineInput(job.details.input)
      : [];

    const inputRepoVertices: Vertex[] = inputs.map(({project, name}) => ({
      id: generateVertexId(project, name),
      project,
      name,
      parents: [],
      type: projectId !== project ? NodeType.CROSS_PROJECT_REPO : NodeType.REPO,
      state: undefined,
      jobState: undefined,
      nodeState: undefined,
      jobNodeState: undefined,
      access: true,
    }));

    const pipelineVertex: Vertex = {
      id: generateVertexId(
        job.job?.pipeline?.project?.name || '',
        job.job?.pipeline?.name || '',
      ),
      project: job.job?.pipeline?.project?.name || '',
      name: job.job?.pipeline?.name || '',
      parents: inputs,
      type: NodeType.PIPELINE,
      access: true,
      state: undefined,
      nodeState: undefined,
      jobState: job.state,
      jobNodeState: restJobStateToNodeState(job.state),
    };

    return [...acc, ...inputRepoVertices, pipelineVertex];
  }, []);

  return uniqBy(vertices, (v) => v.id);
};
