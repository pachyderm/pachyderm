import {Chart as ChartJS} from 'chart.js';
import throttle from 'lodash/throttle';
import React, {useState, useRef, useCallback, useEffect} from 'react';

import styles from './ZoomPanSlider.module.css';

type ZoomPanSliderProps = {
  chartRef: React.RefObject<ChartJS<'bar'>>;
};

const transparentImage = new Image();
transparentImage.src =
  'data:image/gif;base64,R0lGODlhAQABAIAAAAUEBAAAACwAAAAAAQABAAACAkQBADs=';

const triggerChartRender = throttle(
  (
    chartRange: number,
    min: number,
    max: number,
    chartRef: React.RefObject<ChartJS<'bar'>>,
  ) => {
    chartRef?.current?.zoomScale('x', {
      min: (min / 100) * chartRange,
      max: (max / 100) * chartRange,
    });
  },
  50,
);

const ZoomPanSlider: React.FC<ZoomPanSliderProps> = ({chartRef}) => {
  const sliderRef = useRef<HTMLDivElement>(null);
  const [extremes, setExtremes] = useState({min: 0, max: 100});
  const [scaleRange, setScaleRange] = useState<number>(0);
  const [sliderRange, setSliderRange] = useState({
    sliderMin: 0,
    sliderMax: 0,
    sliderWidth: 0,
  });
  const {min, max} = extremes;

  const handleMaxSliderChange = (event: React.DragEvent) => {
    // clientX is currently not supported in firefox due to a bug
    event.preventDefault();
    const xPos =
      ((event.clientX - sliderRange.sliderMin) / sliderRange.sliderWidth) * 100;
    if (xPos <= 100 && xPos > min) {
      setExtremes((extremes) => ({max: xPos, min: extremes.min}));
    }
  };

  const handleMinSliderChange = (event: React.DragEvent) => {
    // clientX is currently not supported in firefox due to a bug
    event.preventDefault();
    const xPos =
      ((event.clientX - sliderRange.sliderMin) / sliderRange.sliderWidth) * 100;
    if (xPos < max && xPos >= 0) {
      setExtremes((extremes) => ({max: extremes.max, min: xPos}));
    }
  };

  const handleRangeSliderMouseMove = useCallback(
    (event: MouseEvent) => {
      event.preventDefault();
      const mouseMove = (event.movementX / sliderRange.sliderWidth) * 100;

      setExtremes((extremes) => {
        const minMove = extremes.min + mouseMove;
        const maxMove = extremes.max + mouseMove;

        if (maxMove <= 100 && minMove >= 0) {
          return {
            max: maxMove <= 100 ? maxMove : 100,
            min: minMove >= 0 ? minMove : 0,
          };
        } else {
          return {
            ...extremes,
          };
        }
      });
    },
    [sliderRange.sliderWidth],
  );

  // We need to assign mousemove/mouseup handlers to the window so the
  // mouse cursor can leave the element while dragging
  const handleRangeSliderMouseDown = () => {
    window.addEventListener('mousemove', handleRangeSliderMouseMove);
  };

  useEffect(() => {
    // record slider element width and position to calculate relative mouse movements
    if (sliderRef) {
      const sliderMin = sliderRef.current?.getBoundingClientRect().x || 0;
      const sliderWidth = sliderRef.current?.getBoundingClientRect().width || 0;
      const sliderMax = sliderWidth + sliderWidth || 0;

      setSliderRange({
        sliderMin,
        sliderMax,
        sliderWidth,
      });
    }

    const cleanupRangeSliderMouseDown = () => {
      window.removeEventListener('mousemove', handleRangeSliderMouseMove);
    };

    // update cursor image while dragging to transparent gif
    const handleDragStart = (event: DragEvent) => {
      event.dataTransfer?.setDragImage(transparentImage, 0, 0);
    };

    window.addEventListener('mouseup', cleanupRangeSliderMouseDown);
    window.addEventListener('dragstart', handleDragStart);
    return () => {
      window.removeEventListener('mouseup', cleanupRangeSliderMouseDown);
      window.removeEventListener('dragstart', handleDragStart);
    };
  }, [sliderRef, handleRangeSliderMouseMove]);

  useEffect(() => {
    const chartRange =
      scaleRange ||
      (chartRef.current?.scales.x.max || 0) -
        (chartRef.current?.scales.x.min || 0);

    if (!scaleRange) {
      // save the initial chart x-axis range to
      // calculate relative min/max values
      setScaleRange(chartRange);
    }

    // update the chart scale based on changing min/max values
    triggerChartRender(chartRange, min, max, chartRef);
  }, [chartRef, max, min, scaleRange]);

  return (
    <div className={styles.base} ref={sliderRef}>
      <div
        className={styles.handle}
        style={{left: `calc(${min}% - 0.5rem)`}}
        draggable
        onDrag={handleMinSliderChange}
      />
      <div
        className={styles.range}
        style={{
          left: `calc(${min}%)`,
          width: `calc(${max - min}%`,
        }}
        onMouseDown={handleRangeSliderMouseDown}
      />
      <div
        className={styles.handle}
        style={{left: `calc(${max}%)`}}
        draggable
        onDrag={handleMaxSliderChange}
      />
    </div>
  );
};

export default ZoomPanSlider;
