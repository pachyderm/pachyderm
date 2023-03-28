import {Job} from '@graphqlTypes';
import {useCallback, useState} from 'react';

const TOOLTIP_OFFSET_X = -20;
const TOOLTIP_OFFSET_Y = 125;

const useRuntimesChartTooltip = (
  jobsCrossReference: Record<string, Record<string, Job>>,
) => {
  const [tooltip, setTooltip] = useState({
    opacity: 0,
    top: 0,
    left: 0,
    runtime: '',
    failedDatums: 0,
  });

  const setTooltipState = useCallback(
    (context) => {
      const tooltipModel = context.tooltip;

      if (tooltipModel.opacity === 0 || !tooltipModel.dataPoints[0]) {
        setTooltip((prev) => ({...prev, opacity: 0}));
        return;
      }
      const data = tooltipModel.dataPoints[0];
      const newTooltipData = {
        opacity: 1,
        left: context.tooltip._eventPosition.x + TOOLTIP_OFFSET_X,
        top: context.tooltip._eventPosition.y + TOOLTIP_OFFSET_Y,
        runtime: String(data.raw[1] - data.raw[0]),
        failedDatums:
          jobsCrossReference[data.dataset.label][data.label].dataFailed,
      };
      setTooltip(newTooltipData);
    },
    [jobsCrossReference],
  );

  return {tooltip, setTooltipState};
};

export default useRuntimesChartTooltip;
