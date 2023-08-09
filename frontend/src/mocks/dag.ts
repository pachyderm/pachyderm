import {
  JobState,
  NodeState,
  NodeType,
  PipelineState,
  mockGetDagQuery,
} from '@graphqlTypes';

const DEFAULT_INPUT_REPO = {
  id: 'default_input_repo',
  name: 'input',
  state: null,
  nodeState: null,
  access: true,
  parents: [],
  type: NodeType.INPUT_REPO,
  jobState: null,
  jobNodeState: null,
  createdAt: 1614126189,
  __typename: 'Vertex' as const,
};
const DEFAULT_MONTAGE_REPO = {
  id: 'default_montage_repo',
  name: 'montage',
  state: null,
  nodeState: null,
  access: true,
  parents: ['default_montage'],
  type: NodeType.OUTPUT_REPO,
  jobState: null,
  jobNodeState: null,
  createdAt: 1614226189,
  __typename: 'Vertex' as const,
};
const DEFAULT_MONTAGE_PIPELINE = {
  id: 'default_montage',
  name: 'montage',
  state: PipelineState.PIPELINE_STATE_UNKNOWN,
  nodeState: NodeState.IDLE,
  access: true,
  parents: ['default_input'],
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
  id: 'default_montage2',
  name: 'montage2',
  state: PipelineState.PIPELINE_STATE_UNKNOWN,
  nodeState: NodeState.IDLE,
  access: true,
  parents: ['default_montage'],
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
