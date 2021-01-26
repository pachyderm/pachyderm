import {NodeType} from 'generated/types';

export interface Context {
  authToken?: string;
  pachdAddress?: string;
}

export type LinkInputData = {
  source: number;
  target: number;
  error?: boolean;
  active?: boolean;
};

export interface Vertex {
  name: string;
  type: NodeType;
  parents: string[];
}
