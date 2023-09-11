import {NodeState, NodeType, JobState, PipelineState} from '@graphqlTypes';
import {ElkExtendedEdge, ElkNode} from 'elkjs/lib/elk-api';
import {CSSProperties} from 'react';

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

export type FileMajorType =
  | 'document'
  | 'image'
  | 'video'
  | 'audio'
  | 'folder'
  | 'unknown';

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
};

export type PointCoordinates = {
  x: number;
  y: number;
};

export type Dag = {
  nodes: Node[];
  links: Link[];
  id: string;
};

export type DagNodes = {
  id: string;
  nodes: Node[];
};
export interface LinkInputData
  extends ElkExtendedEdge,
    Pick<Link, 'state' | 'targetState' | 'sourceState' | 'transferring'> {}

export interface NodeInputData extends ElkNode, Omit<Node, 'x' | 'y'> {}

export type PfsInput = {
  repo: string;
  project: string;
};
export type Input = {
  pfs?: PfsInput;
  joinList: [Input];
  groupList: [Input];
  crossList: [Input];
  unionList: [Input];
};
