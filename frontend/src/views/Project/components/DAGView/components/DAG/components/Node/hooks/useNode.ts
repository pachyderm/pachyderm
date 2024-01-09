import {select} from 'd3-selection';
import {useCallback, useEffect, useMemo} from 'react';
import {useHistory} from 'react-router';

import useLogsNavigation from '@dash-frontend/hooks/useLogsNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {NodeType, Node} from '@dash-frontend/lib/types';
import useDAGRouteController from '@dash-frontend/views/Project/components/DAGView/hooks/useDAGRouteController';
import {NODE_WIDTH} from '@dash-frontend/views/Project/constants/nodeSizes';
import {
  lineageRoute,
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';

const LABEL_WIDTH = NODE_WIDTH - 44;

const setText = (selector: string, groupName: string, elementText: string) => {
  const text = select<SVGGElement, Node>(
    `#${groupName}`,
  ).select<SVGTextElement>(selector);

  // remove old tspans on node name change
  text.selectAll<SVGTSpanElement, unknown>('tspan').remove();

  // create tspans
  const tspan = text.append('tspan');
  const nameChars = elementText.split('').reverse();
  const line: string[] = [];
  const tspanNode = tspan.node();

  const maxWidth = LABEL_WIDTH;

  while (nameChars.length > 0) {
    const char = nameChars.pop();
    if (char) {
      line.push(char);
      tspan.text(line.join(''));

      // If the width of the tspan exceeds maxWidth replace the last 3 characters with '...'
      if (tspanNode && tspanNode.getComputedTextLength() > maxWidth) {
        line.splice(line.length - 3, 3, '...');
        tspan.text(line.join(''));
        break;
      }
    }
  }
};

const useNode = (node: Node, isInteractive: boolean, hideDetails: boolean) => {
  const {
    navigateToNode,
    selectedPipeline,
    selectedRepo,
    pipelinePathMatch,
    repoPathMatch,
  } = useDAGRouteController();
  const {projectId} = useUrlState();
  const {getPathToLatestJobLogs} = useLogsNavigation();
  const browserHistory = useHistory();

  const noAccess = !node.access;

  const groupName = useMemo(() => `GROUP_${node.id}`, [node]);

  const onClick = useCallback(
    (
      destination:
        | 'pipeline'
        | 'repo'
        | 'logs'
        | 'status'
        | 'connected_repo'
        | 'connected_project',
    ) => {
      if (
        ('pipeline' === destination || 'repo' === destination) &&
        isInteractive
      )
        navigateToNode(node, destination);
      if (noAccess) return;
      if (destination === 'logs') {
        return browserHistory.push(
          getPathToLatestJobLogs({
            projectId,
            pipelineId: node.name,
          }),
        );
      }
      if (destination === 'status') {
        return browserHistory.push(
          pipelineRoute({
            projectId,
            pipelineId: node.name,
            tabId: 'info',
          }),
        );
      }

      if (destination === 'connected_project') {
        return browserHistory.push(
          lineageRoute({
            projectId: node.project,
          }),
        );
      }
      if (destination === 'connected_repo') {
        return browserHistory.push(
          repoRoute({
            projectId: node.project,
            repoId: node.name,
          }),
        );
      }
    },
    [
      noAccess,
      isInteractive,
      navigateToNode,
      node,
      browserHistory,
      getPathToLatestJobLogs,
      projectId,
    ],
  );

  useEffect(() => {
    select<SVGGElement, Node>(`#${groupName}`).data([node]);
  }, [groupName, node]);

  useEffect(() => {
    if (hideDetails) return;

    setText('.nodeLabel', groupName, node.name);

    // This type of node has a second line of information we need to display.
    if (node.type === NodeType.CROSS_PROJECT_REPO) {
      setText('.nodeLabelProject', groupName, node.project);
    }
  }, [node, groupName, hideDetails]);

  const repoSelected = selectedRepo === node.name && !!repoPathMatch;

  const pipelineSelected =
    selectedPipeline === node.name && !!pipelinePathMatch;

  return {
    onClick,
    repoSelected,
    pipelineSelected,
    groupName,
  };
};

export default useNode;
