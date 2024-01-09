import flatMap from 'lodash/flatMap';
import keyBy from 'lodash/keyBy';

import {RepoInfo} from '@dash-frontend/api/pfs';
import {PipelineInfo, PipelineState, JobState} from '@dash-frontend/api/pps';
import {
  restJobStateToNodeState,
  restPipelineStateToNodeState,
} from '@dash-frontend/api/utils/nodeStateMappers';
import {getUnixSecondsFromISOString} from '@dash-frontend/lib/dateTime';
import hasRepoReadPermissions from '@dash-frontend/lib/hasRepoReadPermissions';
import {NodeType, NodeState} from '@dash-frontend/lib/types';

import {
  egressNodeName,
  generateVertexId,
  flattenPipelineInput,
  VertexIdentifier,
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
};

const convertPachydermRepoToVertex = ({repo: r}: {repo: RepoInfo}): Vertex => {
  const project = r.repo?.project?.name || '';
  const name = r.repo?.name || '';
  return {
    id: generateVertexId(project, name),
    project,
    name,
    type: NodeType.REPO,
    state: undefined,
    jobState: undefined,
    nodeState: undefined,
    jobNodeState: undefined,
    access: hasRepoReadPermissions(r?.authInfo?.permissions),
    createdAt: getUnixSecondsFromISOString(r.created),
    parents: [],
  };
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
    type: NodeType.PIPELINE,
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
  };
  const vertices: Vertex[] = [pipelineNode];

  const egress = p.details?.egress;
  if (!egress || !egressNodeName(egress)) return vertices;
  // Egress nodes are a visual representation of a pipeline setting that is
  // unique to our DAG, so we want to append a identifier to the name to have it
  // differ from the pipeline name.
  const egressName = `${name}.egress`;
  const egressNode = {
    id: generateVertexId(project, egressName),
    project,
    name: egressNodeName(egress) || '',
    type: NodeType.EGRESS,
    access: true,
    parents: [
      {
        id: generateVertexId(project, name),
        project,
        name,
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

  const repoVertices = filteredRepos.map<Vertex>((repo) =>
    convertPachydermRepoToVertex({repo}),
  );
  const pipelineVertices = flatMap(pipelines, (pipeline) =>
    convertPachydermPipelineToVertex({pipeline, repoMap}),
  );

  const crossProjectProvenanceRepos = findCrossProjectProvenanceRepos(
    projectId,
    pipelineVertices,
  );

  return [...repoVertices, ...pipelineVertices, ...crossProjectProvenanceRepos];
};
