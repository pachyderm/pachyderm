import {useClipboardCopy} from '@pachyderm/components';
import {select} from 'd3-selection';
import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {useRouteMatch} from 'react-router';

import useHoveredNode from '@dash-frontend/providers/HoveredNodeProvider/hooks/useHoveredNode';
import {
  JOB_PATH,
  PIPELINE_JOB_PATH,
} from '@dash-frontend/views/Project/constants/projectPaths';
import {Node, NodeState, NodeType} from '@graphqlTypes';
import useRouteController from 'hooks/useRouteController';
import deriveRepoNameFromNode from 'lib/deriveRepoNameFromNode';

const useNode = (node: Node, isInteractive: boolean) => {
  const {navigateToNode, selectedNode} = useRouteController();
  const match = useRouteMatch(PIPELINE_JOB_PATH);
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
    select<SVGGElement, Node>(`#${groupName}`).data([node]);
  }, [groupName, node]);

  const normalizedNodeName = deriveRepoNameFromNode(node);

  const isHovered = useMemo(
    () => hoveredNode === node.name,
    [hoveredNode, node.name],
  );

  const isJobPath = match?.isExact;

  const nodeIconHref = useMemo(() => {
    if (!node.access) {
      return '/dag_no_access.svg';
    }

    switch (node.state) {
      case NodeState.SUCCESS:
        return '/dag_success.svg';
      case NodeState.BUSY:
        return '/dag_busy.svg';
      case NodeState.ERROR:
        return isJobPath ? '/dag_error.svg' : '/dag_pipeline_error.svg';
      case NodeState.PAUSED:
        return '/dag_paused.svg';
    }
  }, [node.access, node.state, isJobPath]);

  return {
    isHovered,
    onClick,
    onMouseOut,
    onMouseOver,
    selectedNode,
    groupName,
    isEgress,
    normalizedNodeName,
    showSuccess,
    showLeaveJob,
    handleLeaveJobClick,
    closeLeaveJob,
    nodeIconHref,
  };
};

export default useNode;
