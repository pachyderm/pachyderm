import {useClipboardCopy} from '@pachyderm/components';
import {select} from 'd3-selection';
import React, {useCallback, useEffect, useMemo, useState} from 'react';
import {useRouteMatch} from 'react-router';

import useIsViewingJob from '@dash-frontend/hooks/useIsViewingJob';
import useHoveredNode from '@dash-frontend/providers/HoveredNodeProvider/hooks/useHoveredNode';
import {NODE_HEIGHT} from '@dash-frontend/views/Project/constants/nodeSizes';
import {PIPELINE_JOB_PATH} from '@dash-frontend/views/Project/constants/projectPaths';
import {Node, NodeState, NodeType} from '@graphqlTypes';
import useRouteController from 'hooks/useRouteController';
import deriveRepoNameFromNode from 'lib/deriveRepoNameFromNode';

const LABEL_WIDTH = 188;

const useNode = (node: Node, isInteractive: boolean) => {
  const {navigateToNode, selectedNode} = useRouteController();
  const match = useRouteMatch(PIPELINE_JOB_PATH);
  const {hoveredNode, setHoveredNode} = useHoveredNode();
  const [showSuccess, setShowSuccess] = useState(false);
  const {copy, supported, copied, reset} = useClipboardCopy(node.name);
  const [showLeaveJob, setShowLeaveJob] = useState(false);
  const isViewingJob = useIsViewingJob();

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
    if (isViewingJob && !isEgress && node.type !== NodeType.PIPELINE) {
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

  useEffect(() => {
    let tspanCount = 1;
    const text = select<SVGGElement, Node>(
      `#${groupName}`,
    ).select<SVGTextElement>('.nodeLabel');

    // remove old tspans on node name change
    text.selectAll<SVGTSpanElement, unknown>('tspan').remove();

    // create tspans
    let tspan = text
      .append('tspan')
      .attr('x', 36)
      .attr('y', NODE_HEIGHT / 2);
    const normalizedNodeName = deriveRepoNameFromNode(node);
    const nameChars = normalizedNodeName.split('').reverse();
    let line: string[] = [];
    let char: string | undefined = '';
    while (nameChars.length > 0) {
      char = nameChars.pop();
      if (char) {
        line.push(char);
        tspan.text(line.join(''));
        const tspanNode = tspan.node();
        if (tspanNode && tspanNode.getComputedTextLength() > LABEL_WIDTH) {
          line.pop();
          if (tspanCount === 3) {
            line.splice(line.length - 3, 3, '...');
            tspan.text(line.join(''));
            break;
          } else {
            tspan.text(line.join(''));
            line = [char];
            tspan = text
              .append('tspan')
              .text(char)
              .attr('x', 36)
              .attr('y', NODE_HEIGHT / 2);
            tspanCount += 1;
          }
        }
      }
    }

    // adjust tspan positioning
    text
      .selectAll<SVGTSpanElement, unknown>('tspan')
      .attr('dy', (_d, i, nodes) => {
        if (i === 0) return -10 * (nodes.length - 1);
        if (i === 1) {
          if (nodes.length === 2) return 10;
          return 0;
        } else {
          return 20;
        }
      })
      .attr('dx', 0);
  }, [node, groupName]);

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
    showSuccess,
    showLeaveJob,
    handleLeaveJobClick,
    closeLeaveJob,
    nodeIconHref,
  };
};

export default useNode;
