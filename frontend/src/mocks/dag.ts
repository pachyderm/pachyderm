import {
  JobState,
  NodeState,
  NodeType,
  PipelineState,
  mockGetDagQuery,
} from '@graphqlTypes';
import objectHash from 'object-hash';

const DEFAULT_INPUT_REPO = {
  id: objectHash({
    project: 'default',
    name: 'input',
  }),
  project: 'default',
  name: 'input',
  state: null,
  nodeState: null,
  access: true,
  parents: [],
  type: NodeType.REPO,
  jobState: null,
  jobNodeState: null,
  createdAt: 1614126189,
  __typename: 'Vertex' as const,
};
const DEFAULT_MONTAGE_REPO = {
  id: objectHash({
    project: 'default',
    name: 'montage',
  }),
  project: 'default',
  name: 'montage',
  state: null,
  nodeState: null,
  access: true,
  parents: [
    {
      id: objectHash({
        project: 'default',
        name: 'montage',
      }),
      project: 'default',
      name: 'montage',
      __typename: 'VertexIdentifier' as const,
    },
  ],
  type: NodeType.REPO,
  jobState: null,
  jobNodeState: null,
  createdAt: 1614226189,
  __typename: 'Vertex' as const,
};
const DEFAULT_MONTAGE_PIPELINE = {
  id: objectHash({
    project: 'default',
    name: 'montage',
  }),
  project: 'default',
  name: 'montage',
  state: PipelineState.PIPELINE_STATE_UNKNOWN,
  nodeState: NodeState.IDLE,
  access: true,
  parents: [
    {
      id: objectHash({
        project: 'default',
        name: 'input',
      }),
      project: 'default',
      name: 'input',
      __typename: 'VertexIdentifier' as const,
    },
  ],
  type: NodeType.PIPELINE,
  jobState: JobState.JOB_SUCCESS,
  jobNodeState: NodeState.SUCCESS,
  createdAt: null,
  __typename: 'Vertex' as const,
};

export const mockGetDag = () =>
  mockGetDagQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        dag: [
          DEFAULT_INPUT_REPO,
          DEFAULT_MONTAGE_REPO,
          DEFAULT_MONTAGE_PIPELINE,
        ],
      }),
    );
  });

const DEFAULT_DOWNSTREAM_PIPELINE = {
  id: objectHash({
    project: 'default',
    name: 'montage2',
  }),
  project: 'default',
  name: 'montage2',
  state: PipelineState.PIPELINE_STATE_UNKNOWN,
  nodeState: NodeState.IDLE,
  access: true,
  parents: [
    {
      id: objectHash({
        project: 'default',
        name: 'montage',
      }),
      project: 'default',
      name: 'montage',
      __typename: 'VertexIdentifier' as const,
    },
  ],
  type: NodeType.PIPELINE,
  jobState: JobState.JOB_SUCCESS,
  jobNodeState: NodeState.SUCCESS,
  createdAt: null,
  __typename: 'Vertex' as const,
};

export const mockGetLargerDag = () =>
  mockGetDagQuery((_req, res, ctx) => {
    return res(
      ctx.data({
        dag: [
          DEFAULT_INPUT_REPO,
          DEFAULT_MONTAGE_REPO,
          DEFAULT_MONTAGE_PIPELINE,
          DEFAULT_DOWNSTREAM_PIPELINE,
        ],
      }),
    );
  });
