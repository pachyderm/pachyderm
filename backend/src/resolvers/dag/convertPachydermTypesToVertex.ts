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

import {egressNodeName, postfixNameWithRepo} from './helpers';

const convertPachydermRepoToVertex = ({
  repo: r,
  pipelineMap,
}: {
  repo: RepoInfo.AsObject;
  pipelineMap: Record<string, PipelineInfo.AsObject>;
}): Vertex => ({
  id: postfixNameWithRepo(`${r.repo?.project?.name || ''}_${r.repo?.name}`),
  name: r.repo?.name || '',
  type:
    r.repo && pipelineMap[`${r.repo?.project?.name || ''}_${r.repo?.name}`]
      ? NodeType.OUTPUT_REPO
      : NodeType.INPUT_REPO,
  state: null,
  jobState: null,
  nodeState: null,
  jobNodeState: null,
  access: hasRepoReadPermissions(r.authInfo?.permissionsList),
  createdAt: r.created?.seconds,
  // if there is a pipeline with the same name as our repo, it is our only parent
  parents:
    r.repo && pipelineMap[`${r.repo?.project?.name || ''}_${r.repo?.name}`]
      ? [`${r.repo?.project?.name || ''}_${r.repo?.name}`]
      : [],
});

const convertPachydermPipelineToVertex = ({
  pipeline: p,
  repoMap,
}: {
  pipeline: PipelineInfo.AsObject;
  repoMap: Record<string, RepoInfo.AsObject>;
}): Vertex[] => {
  const pipelineName = `${p.pipeline?.project?.name || ''}_${p.pipeline?.name}`;
  const state = toGQLPipelineState(p.state);
  const jobState = toGQLJobState(p.lastJobState);

  const pipelineNode = {
    id: pipelineName,
    name: p.pipeline?.name || '',
    type: NodeType.PIPELINE,
    state: state,
    nodeState: gqlPipelineStateToNodeState(state),
    access: p.pipeline
      ? hasRepoReadPermissions(
          repoMap[`${p.pipeline?.project?.name || ''}_${p.pipeline?.name}`]
            ?.authInfo?.permissionsList,
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

  const egressNode = {
    id: `${p.pipeline?.project?.name || ''}_${p.pipeline?.name || ''}_egress`,
    name: egressNodeName(egress) || '',
    type: NodeType.EGRESS,
    access: true,
    parents: [postfixNameWithRepo(pipelineName)],
    state: null,
    jobState: null,
    nodeState: null,
    jobNodeState: null,
    createdAt: p.details?.createdAt?.seconds,
  };
  vertices.push(egressNode);

  return vertices;
};

export const convertPachydermTypesToVertex = (
  repos: RepoInfo.AsObject[],
  pipelines: PipelineInfo.AsObject[],
) => {
  const pipelineMap = keyBy(
    pipelines,
    (p) => `${p.pipeline?.project?.name || ''}_${p.pipeline?.name}`,
  );
  const repoMap = keyBy(
    repos,
    (r) => `${r.repo?.project?.name || ''}_${r.repo?.name}`,
  );

  const repoVertices = repos.map<Vertex>((repo) =>
    convertPachydermRepoToVertex({repo, pipelineMap}),
  );
  const pipelineVertices = flatMap(pipelines, (pipeline) =>
    convertPachydermPipelineToVertex({pipeline, repoMap}),
  );

  return [...repoVertices, ...pipelineVertices];
};
