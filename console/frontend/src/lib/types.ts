import {Graph} from '@dagrejs/graphlib';
import {ElkExtendedEdge, ElkNode} from 'elkjs/lib/elk-api';
import {CSSProperties} from 'react';

import {
  DatumState,
  Job,
  JobInfo,
  JobState,
  PipelineState,
} from '@dash-frontend/api/pps';

export interface ProjectRouteParams {
  projectId: string;
  repoId?: string;
  pipelineId?: string;
  branchId?: string;
  tabId?: string;
  commitId?: string;
  filePath?: string;
  jobId?: string;
  pipelineJobId?: string;
  datumId?: string;
}

export interface ConsoleRouteParams extends ProjectRouteParams {
  view?: string;
}

export type FixedListRowProps = {
  index: number;
  style: CSSProperties;
};

export type FixedGridRowProps = {
  columnIndex: number;
  rowIndex: number;
  style: CSSProperties;
};

export enum DagDirection {
  DOWN = 'DOWN',
  RIGHT = 'RIGHT',
}

export enum NodeState {
  BUSY = 'BUSY',
  ERROR = 'ERROR',
  IDLE = 'IDLE',
  PAUSED = 'PAUSED',
  RUNNING = 'RUNNING',
  SUCCESS = 'SUCCESS',
}

export enum NodeType {
  CROSS_PROJECT_REPO = 'CROSS_PROJECT_REPO',
  CRON_REPO = 'CRON_REPO',
  EGRESS = 'EGRESS',
  PIPELINE = 'PIPELINE',
  SPOUT_PIPELINE = 'SPOUT_PIPELINE',
  SERVICE_PIPELINE = 'SERVICE_PIPELINE',
  SPOUT_SERVICE_PIPELINE = 'SPOUT_SERVICE_PIPELINE',
  REPO = 'REPO',
}

export enum EgressType {
  S3 = 'S3',
  STORAGE = 'storage',
  DB = 'DB',
}

export type Node = {
  id: string;
  project: string;
  name: string;
  type: NodeType;
  x: number;
  y: number;
  state?: PipelineState;
  jobState?: JobState;
  nodeState?: NodeState;
  jobNodeState?: NodeState;
  access: boolean;
  hasCronInput?: boolean;
  parallelism?: string;
  egressType?: EgressType;
};

export type NodeWithId = {id: string; name: string};
export type InputOutputNodesMap = Record<string, NodeWithId[]>;

export type Link = {
  id: string;
  source: NodeWithId;
  sourceState?: NodeState;
  targetState?: NodeState;
  target: NodeWithId;
  state?: NodeState;
  bendPoints: Array<PointCoordinates>;
  startPoint: PointCoordinates;
  endPoint: PointCoordinates;
  transferring: boolean;
  isCrossProject: boolean;
};

export type PointCoordinates = {
  x: number;
  y: number;
};

export type Dags = {
  nodes: Node[];
  links: Link[];
  graph: Graph;
  reverseGraph: Graph;
};

export interface LinkInputData
  extends ElkExtendedEdge,
    Pick<Link, 'state' | 'targetState' | 'sourceState' | 'transferring'> {}

export interface NodeInputData extends ElkNode, Omit<Node, 'x' | 'y'> {}

export type PfsInput = {
  repo: string;
  project: string;
};

export type InternalJobSet = {
  job: Job | undefined;
  created: string | undefined;
  started: string | undefined;
  finished: string | undefined;
  inProgress: boolean;
  state: JobState | undefined;
  jobs: JobInfo[];
};

export enum FileCommitState {
  ADDED = 'ADDED',
  DELETED = 'DELETED',
  UPDATED = 'UPDATED',
}

export type FormattedFileDiff = {
  diffTotals: Record<string, FileCommitState>;
  diff: {
    size: number;
    sizeDisplay: string;
    filesAdded: {
      count: number;
      sizeDelta: number;
    };
    filesUpdated: {
      count: number;
      sizeDelta: number;
    };
    filesDeleted: {
      count: number;
      sizeDelta: number;
    };
  };
};

export type DatumFilter = Exclude<
  DatumState,
  DatumState.UNKNOWN | DatumState.STARTING
>;
