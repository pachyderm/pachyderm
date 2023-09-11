import flatMap from 'lodash/flatMap';
import keyBy from 'lodash/keyBy';

import flattenPipelineInput from '@dash-backend/lib/flattenPipelineInput';
import {
  toGQLJobState,
  toGQLPipelineState,
} from '@dash-backend/lib/gqlEnumMappers';
import hasRepoReadPermissions from '@dash-backend/lib/hasRepoReadPermissions';
import {
  gqlJobStateToNodeState,
  gqlPipelineStateToNodeState,
} from '@dash-backend/lib/nodeStateMappers';
import {RepoInfo, PipelineInfo} from '@dash-backend/proto';
import {NodeType, Vertex} from '@graphqlTypes';

import {egressNodeName, generateVertexId} from './helpers';

const convertPachydermRepoToVertex = ({
  repo: r,
}: {
  repo: RepoInfo.AsObject;
}): Vertex => {
  const project = r.repo?.project?.name || '';
  const name = r.repo?.name || '';
  return {
    id: generateVertexId(project, name),
    project,
    name,
    type: NodeType.REPO,
    state: null,
    jobState: null,
    nodeState: null,
    jobNodeState: null,
    access: hasRepoReadPermissions(r.authInfo?.permissionsList),
    createdAt: r.created?.seconds,
    parents: [],
  };
};

const convertPachydermPipelineToVertex = ({
  pipeline: p,
  repoMap,
}: {
  pipeline: PipelineInfo.AsObject;
  repoMap: Record<string, RepoInfo.AsObject>;
}): Vertex[] => {
  const project = p.pipeline?.project?.name || '';
  const name = p.pipeline?.name || '';

  const state = toGQLPipelineState(p.state);
  const jobState = toGQLJobState(p.lastJobState);

  const pipelineNode = {
    id: generateVertexId(project, name),
    project,
    name,
    type: NodeType.PIPELINE,
    state: state,
    nodeState: gqlPipelineStateToNodeState(state),
    access: p.pipeline
      ? hasRepoReadPermissions(
          repoMap[`${project}.${name}`]?.authInfo?.permissionsList,
        )
      : false,
    jobState: jobState,
    jobNodeState: gqlJobStateToNodeState(jobState),
    createdAt: p.details?.createdAt?.seconds,
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
    state: null,
    jobState: null,
    nodeState: null,
    jobNodeState: null,
    createdAt: p.details?.createdAt?.seconds,
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
          id: generateVertexId(project, name),
          project,
          name,
          state: null,
          nodeState: null,
          access: true,
          parents: [],
          type: NodeType.CROSS_PROJECT_REPO,
          jobState: null,
          jobNodeState: null,
          createdAt: null,
        };
        crossProjectProvenanceRepos.push(vertex);
      }
    });
  });

  return crossProjectProvenanceRepos;
};

export const convertPachydermTypesToVertex = (
  projectId: string,
  repos: RepoInfo.AsObject[],
  pipelines: PipelineInfo.AsObject[],
) => {
  const pipelineMap = keyBy(
    pipelines,
    (p) => `${p.pipeline?.project?.name || ''}.${p.pipeline?.name}`,
  );
  const repoMap = keyBy(
    repos,
    (r) => `${r.repo?.project?.name || ''}.${r.repo?.name}`,
  );

  //We want to filter out output repos as we do not render them in the DAG.
  const filteredRepos = repos.filter(
    (r) => !(r && pipelineMap[`${r.repo?.project?.name}.${r.repo?.name}`]),
  );

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
