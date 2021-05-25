import {useClipboardCopy} from '@pachyderm/components';
import {select} from 'd3-selection';
import {useCallback, useEffect, useMemo, useState} from 'react';

import useHoveredNode from '@dash-frontend/providers/HoveredNodeProvider/hooks/useHoveredNode';
import {Node, NodeType} from '@graphqlTypes';
import useRouteController from 'hooks/useRouteController';
import deriveRepoNameFromNode from 'lib/deriveRepoNameFromNode';

import convertNodeStateToDagState from '../../../utils/convertNodeStateToDagState';

const useNode = (node: Node, isInteractive: boolean) => {
  const {navigateToNode, selectedNode} = useRouteController();
  const {hoveredNode, setHoveredNode} = useHoveredNode();
  const [showSuccess, setShowSuccess] = useState(false);
  const {copy, supported, copied, reset} = useClipboardCopy(node.name);

  const isEgress = node.type === NodeType.EGRESS;

  const groupName = useMemo(() => {
    let nodeName = node.name;
    // Need to have a string that works as a valid query selector.
    // urls (and url-encoded urls) are not valid query selectors.
    if (node.type === NodeType.EGRESS) {
      const {host} = new URL(node.name);

      nodeName = host.replace('.com', '');
    }

    return `${nodeName}GROUP`;
  }, [node]);

  const onClick = useCallback(() => {
    if (isInteractive) navigateToNode(node);
    if (isInteractive && isEgress && supported) copy();
  }, [node, isInteractive, navigateToNode, isEgress, supported, copy]);

  const onMouseOver = useCallback(() => {
    if (isInteractive) {
      select(`#${groupName}`).raise();
      setHoveredNode(node.name);
    }
  }, [isInteractive, setHoveredNode, node, groupName]);

  const onMouseOut = useCallback(() => {
    if (isInteractive) setHoveredNode('');
  }, [isInteractive, setHoveredNode]);

  useEffect(() => {
    let timeout: NodeJS.Timeout;

    if (copied) {
      setShowSuccess(true);

      timeout = setTimeout(() => {
        setShowSuccess(false);
        reset();
      }, 5000);
    }

    return () => {
      timeout && clearTimeout(timeout);
    };
  }, [copied, reset]);

  const state = convertNodeStateToDagState(node.state);

  const normalizedNodeName = deriveRepoNameFromNode(node);

  return {
    hoveredNode,
    onClick,
    onMouseOut,
    onMouseOver,
    selectedNode,
    state,
    groupName,
    isEgress,
    normalizedNodeName,
    showSuccess,
  };
};

export default useNode;
