import {select} from 'd3-selection';
import {useCallback, useEffect, useMemo} from 'react';
import {useHistory} from 'react-router';

import useLogsNavigation from '@dash-frontend/hooks/useLogsNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {Node} from '@dash-frontend/lib/types';
import useDAGRouteController from '@dash-frontend/views/Project/components/DAGView/hooks/useDAGRouteController';
import {NODE_WIDTH} from '@dash-frontend/views/Project/constants/nodeSizes';
import {pipelineRoute} from '@dash-frontend/views/Project/utils/routes';

const LABEL_WIDTH = NODE_WIDTH - 44;

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
    (destination: 'pipeline' | 'repo' | 'logs' | 'status') => {
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
      if (isInteractive) navigateToNode(node, destination);
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

    const text = select<SVGGElement, Node>(
      `#${groupName}`,
    ).select<SVGTextElement>('.nodeLabel');

    // remove old tspans on node name change
    text.selectAll<SVGTSpanElement, unknown>('tspan').remove();

    // create tspans
    const tspan = text.append('tspan');
    const nameChars = node.name.split('').reverse();
    const line: string[] = [];
    const tspanNode = tspan.node();

    const maxWidth = LABEL_WIDTH;

    while (nameChars.length > 0) {
      const char = nameChars.pop();
      if (char) {
        line.push(char);
        tspan.text(line.join(''));

        if (tspanNode && tspanNode.getComputedTextLength() > maxWidth) {
          line.splice(line.length - 3, 3, '...');
          tspan.text(line.join(''));
          break;
        }
      }
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
