import {select} from 'd3-selection';
import {useCallback} from 'react';

import useHoveredNode from '@dash-frontend/providers/HoveredNodeProvider/hooks/useHoveredNode';
import {Node} from '@graphqlTypes';

import useRouteController from '../../../hooks/useRouteController';
import convertNodeStateToDagState from '../../../utils/convertNodeStateToDagState';

const useNode = (node: Node, isInteractive: boolean) => {
  const {navigateToNode, selectedNode} = useRouteController();
  const {hoveredNode, setHoveredNode} = useHoveredNode();

  const onClick = useCallback(() => {
    if (isInteractive) navigateToNode(node);
  }, [node, isInteractive, navigateToNode]);

  const onMouseOver = useCallback(() => {
    if (isInteractive) {
      select(`#${node.name}GROUP`).raise();
      setHoveredNode(node.name);
    }
  }, [isInteractive, setHoveredNode, node]);

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
  };
};

export default useNode;
