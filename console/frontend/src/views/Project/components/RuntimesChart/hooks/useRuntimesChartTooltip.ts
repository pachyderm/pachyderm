import {useCallback, useState} from 'react';

import {JobInfo} from '@dash-frontend/api/pps';

const TOOLTIP_OFFSET_X = 50;
const TOOLTIP_OFFSET_Y = -150;

type Tooltip = {
  opacity: number;
  top: number;
  left: number;
  runtime: string;
  failedDatums: number;
};

const useRuntimesChartTooltip = (
  jobsCrossReference: Record<string, Record<string, JobInfo>>,
) => {
  const [tooltip, setTooltip] = useState<Tooltip>({
    opacity: 0,
    top: 0,
    left: 0,
    runtime: '',
    failedDatums: 0,
  });

  const setTooltipState = useCallback(
    // TODO: fix this type
    (context: any) => {
      const tooltipModel = context.tooltip;

      if (tooltipModel.opacity === 0 || !tooltipModel.dataPoints[0]) {
        setTooltip((prev) => ({...prev, opacity: 0}));
        return;
      }
      const position = context.chart.canvas.getBoundingClientRect();
      const data = tooltipModel.dataPoints[0];

      const newTooltipData: Tooltip = {
        opacity: 1,
        left: tooltipModel._eventPosition.x + TOOLTIP_OFFSET_X,
        top: tooltipModel._eventPosition.y + position.top + TOOLTIP_OFFSET_Y,
        runtime: String(data.raw[1] - data.raw[0]),
        failedDatums: jobsCrossReference[data.dataset.label][data.label]
          ? Number(
              jobsCrossReference[data.dataset.label][data.label].dataFailed,
            ) ?? 0
          : 0,
      };
      setTooltip(newTooltipData);
    },
    [jobsCrossReference],
  );

  return {tooltip, setTooltipState};
};

export default useRuntimesChartTooltip;
