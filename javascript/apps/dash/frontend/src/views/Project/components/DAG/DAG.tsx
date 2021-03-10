import classnames from 'classnames';
import * as d3 from 'd3';
import React, {useState, useEffect} from 'react';

import {Dag} from '@graphqlTypes';

import styles from './DAG.module.css';
import useDAG from './hooks/useDAG';

type DagProps = {
  id: string;
  nodeWidth: number;
  nodeHeight: number;
  fixedWidth?: number;
  fixedHeight?: number;
  preview?: boolean;
  data: Dag;
};

const DAG: React.FC<DagProps> = ({
  data,
  id,
  nodeWidth,
  nodeHeight,
  fixedHeight,
  fixedWidth,
  preview = false,
}) => {
  const [svgParentSize, setSVGParentSize] = useState({
    width: 0,
    height: 0,
  });
  const [svgOffset, setSvgOffset] = useState({
    x: 0,
    y: 0,
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
  const dagHeight = Math.max(svgParentSize.height, window.innerHeight - 80);
  const dagWidth = Math.max(svgParentSize.width, window.innerWidth - 410);

  useEffect(() => {
    const svgElement = d3.select<SVGSVGElement, null>(`#${id}`).node();
    const padding = 30;
    if (svgElement) {
      setSvgOffset({
        x: (svgElement?.getBBox()?.x || 0) - padding,
        y: (svgElement?.getBBox()?.y || 0) - padding,
      });
    }
  }, [id, svgParentSize, nodeHeight]);

  return (
    <div
      className={classnames(styles.base, {[styles.draggable]: !preview})}
      id={`${id}Base`}
    >
      <svg
        id={id}
        preserveAspectRatio="xMaxYMax meet"
        viewBox={`${svgOffset.x} ${svgOffset.y} ${dagWidth} ${dagHeight}`}
        width={fixedWidth || dagWidth}
        height={fixedHeight || dagHeight}
        className={styles.parent}
      />
    </div>
  );
};

export default DAG;
