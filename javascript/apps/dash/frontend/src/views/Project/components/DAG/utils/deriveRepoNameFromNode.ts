import {Node, NodeType} from '@graphqlTypes';

const deriveRouteParamFromNode = (n: Node) => {
  if (n.type === NodeType.Repo) {
    return n.name.replace('_repo', '');
  }
  return n.name;
};

export default deriveRouteParamFromNode;
