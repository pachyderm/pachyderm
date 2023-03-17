import {Chart as ChartJS, Tooltip as TooltipChartJS} from 'chart.js';
import React from 'react';

import {CaptionTextSmall} from '@pachyderm/components';

import styles from './Tooltip.module.css';

ChartJS.register(TooltipChartJS);

type TooltipProps = {
  tooltipState: {
    top: number;
    left: number;
    opacity: number;
    runtime: string;
  };
};

const Tooltip: React.FC<TooltipProps> = ({tooltipState}) => {
  return (
    <div
      className={styles.base}
      style={{
        top: tooltipState.top,
        left: tooltipState.left,
        opacity: tooltipState.opacity,
      }}
    >
      <div className={styles.tooltipLine}>
        <CaptionTextSmall>{tooltipState.runtime} seconds</CaptionTextSmall>
        <CaptionTextSmall>Runtime</CaptionTextSmall>
      </div>
      <CaptionTextSmall className={styles.tooltipText}>
        Click to view datums
      </CaptionTextSmall>
    </div>
  );
};

export default Tooltip;
