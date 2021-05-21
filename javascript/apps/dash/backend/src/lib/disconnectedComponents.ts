import keyBy from 'lodash/keyBy';
import mapValues from 'lodash/mapValues';

import {Link, Node} from '@graphqlTypes';

const disconnectedComponents = (nodes: Node[], links: Link[]) => {
  const nodeByName = keyBy(nodes, (n) => n.name);
  const isVisited = mapValues(nodeByName, () => false);

  const dfs = (node: Node) => {
    isVisited[node.name] = true;

    const result: {nodes: Node[]; links: Link[]} = {
      nodes: [node],
      links: [],
    };

    links.forEach((link) => {
      if (link.source === node.name) result.links.push(link);

      if (link.source === node.name && !isVisited[link.target]) {
        const target = nodes.find((node) => node.name === link.target);
        if (target) {
          const dfsResult = dfs(target);
          result.nodes.push(...dfsResult.nodes);
          result.links.push(...dfsResult.links);
        }
      }
      if (link.target === node.name && !isVisited[link.source]) {
        const source = nodes.find((node) => node.name === link.source);
        if (source) {
          const dfsResult = dfs(source);
          result.nodes.push(...dfsResult.nodes);
          result.links.push(...dfsResult.links);
        }
      }
    });

    return result;
  };

  const components: {nodes: Node[]; links: Link[]}[] = [];

  nodes.forEach((node) => {
    if (!isVisited[node.name]) {
      components.push(dfs(node));
    }
  });

  return components;
};

export default disconnectedComponents;
