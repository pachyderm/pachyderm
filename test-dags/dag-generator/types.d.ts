declare module 'random-dag' {
  import type {Graph} from 'graphlib';

  export function graphlib(
    opts: {
      max_per_rank: number; // how 'fat' the DAG should be
      min_per_rank: number;
      max_ranks: number; // how 'tall' the DAG should be
      min_ranks: number;
      probability: number; // chance of having an edge
    },
    callback: (e: Error, graph: Graph) => void,
  ) {
    throw new Error('Function not implemented.');
  }
}
