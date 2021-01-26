import keyBy from 'lodash/keyBy';
import mapValues from 'lodash/mapValues';

import {Vertex} from './types';

const disconnectedComponents = (nodes: Vertex[]) => {
  const nodeByName = keyBy(nodes, (n) => n.name);
  const isVisited = mapValues(nodeByName, () => false);
  const adjNodesByNodeName = mapValues(nodeByName, () => new Set<Vertex>());

  nodes.forEach((node) => {
    node.parents.forEach((parentName) => {
      const parentNode = nodeByName[parentName];

      if (parentNode) {
        adjNodesByNodeName[node.name].add(parentNode);
        adjNodesByNodeName[parentNode.name].add(node);
      }
    });
  });

  const dfs = (node: Vertex) => {
    isVisited[node.name] = true;

    const result = [node];

    adjNodesByNodeName[node.name].forEach((adj) => {
      if (!isVisited[adj.name]) {
        result.push(...dfs(adj));
      }
    });

    return result;
  };

  const components: Vertex[][] = [];

  nodes.forEach((node) => {
    if (!isVisited[node.name]) {
      components.push(dfs(node));
    }
  });

  return components;
};

export default disconnectedComponents;
