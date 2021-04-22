import {Group, Circle, Tooltip} from '@pachyderm/components';
import classnames from 'classnames';
import * as d3 from 'd3';
import noop from 'lodash/noop';
import React, {useState} from 'react';

import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import {Dag} from '@graphqlTypes';

import styles from './DAG.module.css';
import useDAG from './hooks/useDAG';
import useRouteController from './hooks/useRouteController';

type DagProps = {
  count: number;
  data: Dag;
  id: string;
  nodeWidth: number;
  nodeHeight: number;
  isInteractive?: boolean;
  setLargestDagWidth: React.Dispatch<React.SetStateAction<number | null>>;
  largestDagWidth: number | null;
};

const DAG: React.FC<DagProps> = ({
  count,
  data,
  id,
  nodeWidth,
  nodeHeight,
  isInteractive = true,
  setLargestDagWidth,
  largestDagWidth,
}) => {
  const [svgParentSize, setSVGParentSize] = useState({
    width: 0,
    height: 0,
  });

  const {navigateToDag} = useDAG({
    id,
    svgParentSize,
    setSVGParentSize,
    nodeWidth,
    nodeHeight,
    data,
    isInteractive,
    setLargestDagWidth,
    largestDagWidth,
    dagCount: count,
  });

  const parent = d3.select<HTMLTableRowElement, null>(`#${id}Base`);
  const svgElement = d3.select<SVGSVGElement, null>(`#${id}`).node();

  const dagWidth = svgElement ? svgElement.getBBox().width + nodeWidth * 2 : 0;
  const dagHeight = svgElement
    ? svgElement.getBBox().height + nodeHeight * 2
    : 0;

  const parentWidth = parent.node()?.clientWidth || 0;
  const height =
    count > 1 ? dagHeight : Math.max(300, window.innerHeight - 160);

  return (
    <div
      id={`${id}Base`}
      className={classnames(styles.base, {
        [styles.draggable]: isInteractive,
      })}
      onClick={!isInteractive ? () => navigateToDag(data.id) : noop}
    >
      {data.priorityPipelineState && (
        <Tooltip tooltipText={data.priorityPipelineState} tooltipKey="status">
          <span className={styles.pipelineStatus}>
            <Group spacing={8} align="center">
              <Circle color="red" />
              {readablePipelineState(data.priorityPipelineState)}
            </Group>
          </span>
        </Tooltip>
      )}
      <svg
        id={id}
        preserveAspectRatio="xMinYMid meet"
        viewBox={`0 0 ${dagWidth} ${dagHeight}`}
        className={styles.parent}
        width={parentWidth}
        height={height}
        style={{
          zoom:
            count !== 1 && largestDagWidth && dagWidth !== largestDagWidth
              ? parentWidth / largestDagWidth
              : undefined,
        }}
      >
        <g id={`${id}Graph`} />
      </svg>
    </div>
  );
};

export default DAG;
