import flatMap from 'lodash/flatMap';
import keyBy from 'lodash/keyBy';

import {RepoInfo} from '@dash-frontend/api/pfs';
import {
  PipelineInfo,
  PipelineState,
  JobState,
  PipelineInfoPipelineType,
} from '@dash-frontend/api/pps';
import {
  restJobStateToNodeState,
  restPipelineStateToNodeState,
} from '@dash-frontend/api/utils/nodeStateMappers';
import {getUnixSecondsFromISOString} from '@dash-frontend/lib/dateTime';
import hasRepoReadPermissions from '@dash-frontend/lib/hasRepoReadPermissions';
import {NodeType, NodeState, EgressType} from '@dash-frontend/lib/types';

import {checkCronInputs} from '../checkCronInputs';

import {
  egressNodeName,
  generateVertexId,
  flattenPipelineInput,
  VertexIdentifier,
  VertexIdentifierType,
  egressType,
} from './DAGhelpers';

export type Vertex = {
  __typename?: 'Vertex';
  access: boolean;
  createdAt?: number;
  id: string;
  jobNodeState?: NodeState;
  jobState?: JobState;
  name: string;
  nodeState?: NodeState;
  parents: Array<VertexIdentifier>;
  project: string;
  state?: PipelineState;
  type: NodeType;
  hasCronInput?: boolean;
  parallelism?: string;
  egressType?: EgressType;
};

const convertPachydermRepoToVertex = ({
  repo: r,
  cronInputs,
}: {
  repo: RepoInfo;
  cronInputs: string[];
}): Vertex => {
  const project = r.repo?.project?.name || '';
  const name = r.repo?.name || '';
  const id = generateVertexId(project, name);
  return {
    id,
    project,
    name,
    type:
      cronInputs.length !== 0 && cronInputs.find((cron_id) => cron_id === id)
        ? NodeType.CRON_REPO
        : NodeType.REPO,
    state: undefined,
    jobState: undefined,
    nodeState: undefined,
    jobNodeState: undefined,
    access: hasRepoReadPermissions(r?.authInfo?.permissions),
    createdAt: getUnixSecondsFromISOString(r.created),
    parents: [],
  };
};

const derivePipelineType = (pipeline: PipelineInfo) => {
  if (pipeline.type === PipelineInfoPipelineType.PIPELINE_TYPE_SPOUT) {
    return pipeline.details?.spout?.service
      ? NodeType.SPOUT_SERVICE_PIPELINE
      : NodeType.SPOUT_PIPELINE;
  } else if (pipeline.type === PipelineInfoPipelineType.PIPELINE_TYPE_SERVICE) {
    return NodeType.SERVICE_PIPELINE;
  }
  return NodeType.PIPELINE;
};

const convertPachydermPipelineToVertex = ({
  pipeline: p,
  repoMap,
}: {
  pipeline: PipelineInfo;
  repoMap: Record<string, RepoInfo>;
}): Vertex[] => {
  const project = p.pipeline?.project?.name || '';
  const name = p.pipeline?.name || '';
  const pipelineNode = {
    id: generateVertexId(project, name),
    project,
    name,
    type: derivePipelineType(p),
    state: p.state,
    nodeState: restPipelineStateToNodeState(p.state),
    access: p.pipeline
      ? hasRepoReadPermissions(
          repoMap[`${project}.${name}`]?.authInfo?.permissions,
        )
      : false,
    jobState: p.lastJobState,
    jobNodeState: restJobStateToNodeState(p.lastJobState),
    createdAt: getUnixSecondsFromISOString(p.details?.createdAt),
    // our parents are any repos found in our input spec
    parents: p.details?.input ? flattenPipelineInput(p.details?.input) : [],
    hasCronInput: checkCronInputs(p.details?.input),
    parallelism: p.parallelism,
  };
  const vertices: Vertex[] = [pipelineNode];

  if (!egressNodeName(p.details?.egress) && !p.details?.s3Out) return vertices;
  // Egress nodes are a visual representation of a pipeline setting that is
  // unique to our DAG, so we want to append a identifier to the name to have it
  // differ from the pipeline name.

  const egressName = p.details?.s3Out
    ? `s3://${name}`
    : egressNodeName(p.details?.egress);

  const egressNode = {
    id: generateVertexId(project, `${name}.egress`),
    project,
    name: egressName || '',
    type: NodeType.EGRESS,
    egressType: egressType(p.details),
    access: true,
    parents: [
      {
        id: generateVertexId(project, name),
        project,
        name,
        type: VertexIdentifierType.DEFAULT,
      },
    ],
    state: undefined,
    jobState: undefined,
    nodeState: undefined,
    jobNodeState: undefined,
    createdAt: getUnixSecondsFromISOString(p.details?.createdAt),
  };
  vertices.push(egressNode);

  return vertices;
};

export const findCrossProjectProvenanceRepos = (
  projectId: string,
  pipelines: Vertex[],
): Vertex[] => {
  const crossProjectProvenanceRepos: Vertex[] = [];

  pipelines.forEach((pipeline) => {
    pipeline.parents.forEach(({project, name}) => {
      if (project !== projectId) {
        const vertex: Vertex = {
          id: generateVertexId(project || '', name || ''),
          project: project || '',
          name: name || '',
          state: undefined,
          nodeState: undefined,
          access: true,
          parents: [],
          type: NodeType.CROSS_PROJECT_REPO,
          jobState: undefined,
          jobNodeState: undefined,
          createdAt: undefined,
        };
        crossProjectProvenanceRepos.push(vertex);
      }
    });
  });

  return crossProjectProvenanceRepos;
};

const findCronInputs = (vertices: Vertex[]): string[] => {
  const cronInputs: string[] = [];
  vertices.forEach((vertex) => {
    vertex.parents.forEach((parent) => {
      if (parent.type === VertexIdentifierType.CRON) {
        cronInputs.push(parent.id);
      }
    });
  });
  return cronInputs;
};

export const convertPachydermTypesToVertex = async (
  projectId: string,
  repos?: RepoInfo[],
  pipelines?: PipelineInfo[],
) => {
  const pipelineMap = keyBy(
    pipelines,
    (p) => `${p.pipeline?.project?.name || ''}.${p.pipeline?.name}`,
  );
  const repoMap = keyBy(
    repos,
    (r) => `${r?.repo?.project?.name || ''}.${r.repo?.name}`,
  );

  //We want to filter out output repos as we do not render them in the DAG.
  const filteredRepos =
    repos?.filter(
      (r) => !(r && pipelineMap[`${r.repo?.project?.name}.${r.repo?.name}`]),
    ) || [];

  const pipelineVertices = flatMap(pipelines, (pipeline) =>
    convertPachydermPipelineToVertex({pipeline, repoMap}),
  );

  const cronInputs = findCronInputs(pipelineVertices);

  const repoVertices = filteredRepos.map<Vertex>((repo) =>
    convertPachydermRepoToVertex({repo, cronInputs}),
  );

  const crossProjectProvenanceRepos = findCrossProjectProvenanceRepos(
    projectId,
    pipelineVertices,
  );

  return [...repoVertices, ...pipelineVertices, ...crossProjectProvenanceRepos];
};
