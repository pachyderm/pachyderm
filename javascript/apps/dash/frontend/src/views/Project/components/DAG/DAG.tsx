import classnames from 'classnames';
import * as d3 from 'd3';
import React, {useState, useEffect} from 'react';

import {Dag} from '@graphqlTypes';

import styles from './DAG.module.css';
import useDAG from './hooks/useDAG';

type DagProps = {
  count: number;
  data: Dag;
  id: string;
  fixedWidth?: number;
  fixedHeight?: number;
  nodeWidth: number;
  nodeHeight: number;
  preview?: boolean;
};

const DAG: React.FC<DagProps> = ({
  count,
  data,
  id,
  fixedHeight,
  fixedWidth,
  nodeWidth,
  nodeHeight,
  preview = false,
}) => {
  const [svgParentSize, setSVGParentSize] = useState({
    width: 0,
    height: 0,
  });

  useDAG({
    id,
    svgParentSize,
    setSVGParentSize,
    nodeWidth,
    nodeHeight,
    data,
    preview,
  });
  // The constants here are static sizes on the page. When the UI gets developed further we can adjust these
  const dagHeight = fixedHeight
    ? fixedHeight
    : Math.max(300, window.innerHeight / count - 120);
  const dagWidth = fixedWidth
    ? fixedWidth
    : Math.max(svgParentSize.width, window.innerWidth - 80);

  useEffect(() => {
    const parent = d3.select<HTMLTableRowElement, null>(`#${id}Base`);
    const width = parent.node()?.clientWidth || 0;
    const height = parent.node()?.clientHeight || 0;
    setSVGParentSize({width, height});
  }, [id]);

  return (
    <div
      id={`${id}Base`}
      className={classnames(styles.base, {
        [styles.draggable]: !preview,
      })}
    >
      <svg
        id={id}
        preserveAspectRatio="xMaxYMax meet"
        viewBox={`0 0 ${dagWidth} ${dagHeight}`}
        className={styles.parent}
        width={fixedWidth}
        height={fixedHeight}
      >
        <g id={`${id}Graph`} />
      </svg>
    </div>
  );
};

export default DAG;
