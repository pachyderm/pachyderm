import {Chart as ChartJS, Tooltip as TooltipChartJS} from 'chart.js';
import React from 'react';

import {CaptionTextSmall} from '@pachyderm/components';

import styles from './Tooltip.module.css';

ChartJS.register(TooltipChartJS);
const TOOLTIP_HEIGHT = 99;
const TOOLTIP_DATUMS_HEIGHT = 137;

type TooltipProps = {
  tooltipState: {
    top: number;
    left: number;
    opacity: number;
    runtime: string;
    failedDatums: number;
  };
  useHoursAsUnit: boolean;
};

const Tooltip: React.FC<TooltipProps> = ({tooltipState, useHoursAsUnit}) => {
  return (
    <div
      className={styles.base}
      style={{
        top:
          tooltipState.top -
          (tooltipState.failedDatums > 0
            ? TOOLTIP_DATUMS_HEIGHT
            : TOOLTIP_HEIGHT),
        left: tooltipState.left,
        opacity: tooltipState.opacity,
      }}
    >
      {tooltipState.failedDatums > 0 && (
        <div className={styles.tooltipLine}>
          <CaptionTextSmall>{tooltipState.failedDatums}</CaptionTextSmall>
          <CaptionTextSmall>Failed Datums</CaptionTextSmall>
        </div>
      )}
      <div className={styles.tooltipLine}>
        <CaptionTextSmall>
          {parseFloat(tooltipState.runtime).toFixed(1)}{' '}
          {useHoursAsUnit ? 'hours' : 'seconds'}
        </CaptionTextSmall>
        <CaptionTextSmall>Runtime</CaptionTextSmall>
      </div>
      <CaptionTextSmall className={styles.tooltipText}>
        Click to view datums
      </CaptionTextSmall>
    </div>
  );
};

export default Tooltip;
