import {useClipboardCopy} from '@pachyderm/components';
import {select} from 'd3-selection';
import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {useRouteMatch} from 'react-router';

import useHoveredNode from '@dash-frontend/providers/HoveredNodeProvider/hooks/useHoveredNode';
import {JOB_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {Node, NodeType} from '@graphqlTypes';
import useRouteController from 'hooks/useRouteController';
import deriveRepoNameFromNode from 'lib/deriveRepoNameFromNode';

import convertNodeStateToDagState from '../../../utils/convertNodeStateToDagState';

const useNode = (
  node: Node,
  isInteractive: boolean,
  offset: {x: number; y: number},
) => {
  const {navigateToNode, selectedNode} = useRouteController();
  const {hoveredNode, setHoveredNode} = useHoveredNode();
  const [showSuccess, setShowSuccess] = useState(false);
  const {copy, supported, copied, reset} = useClipboardCopy(node.name);
  const [showLeaveJob, setShowLeaveJob] = useState(false);
  const isViewingJob = useRouteMatch(JOB_PATH);

  const isEgress = node.type === NodeType.EGRESS;
  const noAccess = !node.access;

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
    if (isViewingJob && !isEgress) {
      setShowLeaveJob(true);
      return;
    }
    if (noAccess) return;
    if (isInteractive) navigateToNode(node);
    if (isInteractive && isEgress && supported) copy();
  }, [
    node,
    isInteractive,
    navigateToNode,
    isEgress,
    supported,
    copy,
    noAccess,
    isViewingJob,
  ]);

  const onMouseOver = useCallback(() => {
    if (isInteractive) {
      select(`#${groupName}`).raise();
      setHoveredNode(node.name);
    }
  }, [isInteractive, setHoveredNode, node, groupName]);

  const onMouseOut = useCallback(() => {
    if (isInteractive) setHoveredNode('');
  }, [isInteractive, setHoveredNode]);

  const closeLeaveJob = useCallback(() => {
    setShowLeaveJob(false);
  }, [setShowLeaveJob]);

  const handleLeaveJobClick = useCallback(
    (e: React.MouseEvent) => {
      // prevent click event from bubbling to node group
      e.stopPropagation();

      setShowLeaveJob(false);
      navigateToNode(node);
    },
    [node, navigateToNode],
  );

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

  useEffect(() => {
    select<SVGGElement, Node>(`#${groupName}`).data([
      {...node, x: node.x + offset.x, y: node.y + offset.y},
    ]);
  }, [groupName, node, offset]);

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
    showLeaveJob,
    handleLeaveJobClick,
    closeLeaveJob,
  };
};

export default useNode;
