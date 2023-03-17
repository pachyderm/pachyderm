import {useCallback, useState} from 'react';

const TOOLTIP_OFFSET_X = -20;
const TOOLTIP_OFFSET_Y = 125;

const useRuntimesChartTooltip = () => {
  const [tooltip, setTooltip] = useState({
    opacity: 0,
    top: 0,
    left: 0,
    runtime: '',
  });

  const setTooltipState = useCallback(
    (context) => {
      const tooltipModel = context.tooltip;

      if (tooltipModel.opacity === 0) {
        if (tooltip.opacity !== 0)
          setTooltip((prev) => ({...prev, opacity: 0}));
        return;
      }
      const newTooltipData = {
        opacity: 1,
        left: context.tooltip._eventPosition.x + TOOLTIP_OFFSET_X,
        top: context.tooltip._eventPosition.y + TOOLTIP_OFFSET_Y,
        runtime: String(
          tooltipModel.dataPoints[0].raw[1] - tooltipModel.dataPoints[0].raw[0],
        ),
      };
      if (tooltip.opacity !== newTooltipData.opacity) {
        setTooltip(newTooltipData);
      }
    },
    [tooltip],
  );

  return {tooltip, setTooltipState};
};

export default useRuntimesChartTooltip;
