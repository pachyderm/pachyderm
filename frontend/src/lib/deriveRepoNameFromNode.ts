import {NodeType, Vertex} from '@graphqlTypes';

import {Node} from './types';

export const deriveNameFromNodeNameAndType = (name: string, type: NodeType) => {
  if (type === NodeType.OUTPUT_REPO || type === NodeType.INPUT_REPO) {
    return name.replace('_repo', '');
  }
  return name;
};

const deriveRepoNameFromNode = (n: Vertex | Node) => {
  return deriveNameFromNodeNameAndType(n.name, n.type);
};

export default deriveRepoNameFromNode;
