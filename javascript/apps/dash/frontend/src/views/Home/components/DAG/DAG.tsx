import React, {useState} from 'react';

import {DataProps} from 'lib/DAGTypes';

import styles from './DAG.module.css';
import useDAG from './hooks/useDAG';

type DagProps = {
  id: string;
  nodeWidth: number;
  nodeHeight: number;
  data: DataProps;
};

const DAG: React.FC<DagProps> = ({data, id, nodeWidth, nodeHeight}) => {
  const [svgParentSize, setSVGParentSize] = useState({
    width: 2000,
    height: 2000,
  });
  useDAG({
    id,
    svgParentSize,
    setSVGParentSize,
    nodeWidth,
    nodeHeight,
    data,
  });
  const dagHeight = Math.max(svgParentSize.height, window.innerHeight - 100);
  const dagWidth = Math.max(svgParentSize.width, window.innerWidth * 0.7 - 100);

  return (
    <div className={styles.base} id={`${id}Base`}>
      {/* preserveAspectRatio and viewbox are necessary to keep the svg contents visible with overflow in some cases */}
      <svg
        id={id}
        preserveAspectRatio="xMinYMin meet"
        viewBox={`0 0 ${svgParentSize.width} ${svgParentSize.height}`}
        width={dagWidth}
        height={dagHeight}
        className={styles.parent}
      />
    </div>
  );
};

export default DAG;
