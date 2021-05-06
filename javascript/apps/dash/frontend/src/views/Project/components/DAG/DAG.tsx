import {Group, Tooltip} from '@pachyderm/components';
import classnames from 'classnames';
import noop from 'lodash/noop';
import React, {useState} from 'react';

import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import HoveredNodeProvider from '@dash-frontend/providers/HoveredNodeProvider';
import {Dag, DagDirection} from '@graphqlTypes';

import {NODE_HEIGHT, NODE_WIDTH} from '../../constants/nodeSizes';

import Link from './components/Link';
import Node from './components/Node';
import RangeSlider from './components/RangeSlider';
import {ReactComponent as RotateSvg} from './components/Rotate.svg';
import styles from './DAG.module.css';
import useDag from './hooks/useDag';

type DagProps = {
  dagsToShow: number;
  data: Dag;
  id: string;
  nodeWidth: number;
  nodeHeight: number;
  isInteractive?: boolean;
  setLargestDagScale: React.Dispatch<React.SetStateAction<number | null>>;
  largestDagScale: number | null;
  rotateDag: () => void;
  dagDirection: DagDirection;
};

const MARKERS = [
  {id: 'end-arrow', color: '#000'},
  {id: 'end-arrow-active', color: '#5ba3b1'},
  {id: 'end-arrow-error', color: '#E02020'},
];

const DAG: React.FC<DagProps> = ({
  dagsToShow,
  data,
  id,
  nodeWidth,
  nodeHeight,
  isInteractive = true,
  setLargestDagScale,
  largestDagScale,
  rotateDag,
  dagDirection,
}) => {
  const [sliderZoomValue, setSliderZoomValue] = useState(1);
  const {navigateToDag, svgSize, applySliderZoom} = useDag({
    data,
    id,
    isInteractive,
    largestDagScale,
    nodeHeight,
    nodeWidth,
    setLargestDagScale,
    setSliderZoomValue,
    dagDirection,
  });
  const isMultiDag = dagsToShow > 1;

  return (
    <HoveredNodeProvider>
      <div
        id={`${id}Base`}
        className={classnames(styles.base, {
          [styles.draggable]: isInteractive,
          [styles.multiDag]: isMultiDag,
        })}
        onClick={!isInteractive ? () => navigateToDag(data.id) : noop}
      >
        {isMultiDag && data.priorityPipelineState && (
          <Tooltip tooltipText={data.priorityPipelineState} tooltipKey="status">
            <span className={styles.pipelineStatus}>
              <Group spacing={8} align="center">
                <img
                  src="/dag_pipeline_error.svg"
                  className={styles.pipelineStatusIcon}
                  alt="Pipeline Error"
                />
                {readablePipelineState(data.priorityPipelineState)}
              </Group>
            </span>
          </Tooltip>
        )}
        {!isMultiDag && (
          <div className={styles.canvasControls}>
            <RangeSlider
              min="60"
              max="100"
              handleChange={(d: React.ChangeEvent<HTMLInputElement>) =>
                applySliderZoom(d)
              }
              value={sliderZoomValue * 100}
            />
            <button
              className={classnames(styles.rotate, [styles[dagDirection]])}
              onClick={rotateDag}
            >
              <RotateSvg
                aria-label={'Rotate Dag'}
                className={styles.rotateSvg}
              />
            </button>
          </div>
        )}
        <svg
          id={id}
          preserveAspectRatio="xMinYMid meet"
          viewBox={`0 0 ${svgSize.width} ${svgSize.height}`}
        >
          <defs>
            {MARKERS.map((marker) => (
              <marker
                key={marker.id}
                viewBox="0 -5 10 10"
                refX={9}
                markerWidth={5}
                markerHeight={5}
                orient="auto"
                id={marker.id}
              >
                <path d="M0,-5L10,0L0,5" fill={marker.color} />
              </marker>
            ))}
          </defs>
          <g id={`${id}Graph`} className={styles.graph}>
            {/* Ordering of links and nodes in DOM is important so nodes are on top layer */}
            {data.links.map((link) => (
              <Link key={link.id} link={link} isInteractive={isInteractive} />
            ))}
            {data.nodes.map((node) => (
              <Node
                key={node.name}
                node={node}
                nodeHeight={NODE_HEIGHT}
                nodeWidth={NODE_WIDTH}
                isInteractive={isInteractive}
              />
            ))}
          </g>
        </svg>
      </div>
    </HoveredNodeProvider>
  );
};

export default DAG;
