import {select} from 'd3-selection';
import {useCallback, useMemo} from 'react';

import useHoveredNode from '@dash-frontend/providers/HoveredNodeProvider/hooks/useHoveredNode';
import {Node, NodeType} from '@graphqlTypes';

import useRouteController from '../../../hooks/useRouteController';
import convertNodeStateToDagState from '../../../utils/convertNodeStateToDagState';

const useNode = (node: Node, isInteractive: boolean) => {
  const {navigateToNode, selectedNode} = useRouteController();
  const {hoveredNode, setHoveredNode} = useHoveredNode();

  const groupName = useMemo(() => {
    let nodeName = node.name;
    // Need to have a string that works as a valid query selector.
    // urls (and url-encoded urls) are not valid query selectors.
    if (node.type === NodeType.EGRESS) {
      const {host} = new URL(node.name);

      nodeName = host.replace('.com', '');
    }

    return `#${nodeName}GROUP`;
  }, [node]);

  const onClick = useCallback(() => {
    if (isInteractive) navigateToNode(node);
  }, [node, isInteractive, navigateToNode]);

  const onMouseOver = useCallback(() => {
    if (isInteractive) {
      select(groupName).raise();
      setHoveredNode(node.name);
    }
  }, [isInteractive, setHoveredNode, node, groupName]);

  const onMouseOut = useCallback(() => {
    if (isInteractive) setHoveredNode('');
  }, [isInteractive, setHoveredNode]);

  const state = convertNodeStateToDagState(node.state);

  return {
    hoveredNode,
    onClick,
    onMouseOut,
    onMouseOver,
    selectedNode,
    state,
    groupName,
  };
};

export default useNode;
