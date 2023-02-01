import keyBy from 'lodash/keyBy';
import mapValues from 'lodash/mapValues';

import {Link, Node} from './types';

const disconnectedComponents = (nodes: Node[], links: Link[]) => {
  const nodeByName = keyBy(nodes, (n) => n.id);
  const isVisited = mapValues(nodeByName, () => false);

  const dfs = (node: Node) => {
    isVisited[node.id] = true;

    const result: {nodes: Node[]; links: Link[]} = {
      nodes: [node],
      links: [],
    };

    links.forEach((link) => {
      if (link.source.id === node.id) result.links.push(link);

      if (link.source.id === node.id && !isVisited[link.target.id]) {
        const target = nodes.find((node) => node.id === link.target.id);
        if (target) {
          const dfsResult = dfs(target);
          result.nodes.push(...dfsResult.nodes);
          result.links.push(...dfsResult.links);
        }
      }
      if (link.target.id === node.id && !isVisited[link.source.id]) {
        const source = nodes.find((node) => node.id === link.source.id);
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
    if (!isVisited[node.id]) {
      components.push(dfs(node));
    }
  });

  return components;
};

export default disconnectedComponents;
