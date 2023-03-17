import {Chart as ChartJS, ChartOptions, ChartData} from 'chart.js';
import React from 'react';
import {Bar} from 'react-chartjs-2';

import styles from './BarChart.module.css';

const Y_AXIS_HEIGHT = 200;

type BarChartProps = {
  chartData: ChartData<'bar'>;
  options: ChartOptions<'bar'>;
  chartRef: React.RefObject<ChartJS<'bar'>>;
  pipelinesLength: number;
  handleBarClick: (
    event: React.MouseEvent<HTMLCanvasElement, MouseEvent>,
  ) => void;
};

const BarChart: React.FC<BarChartProps> = ({
  chartData,
  options,
  chartRef,
  pipelinesLength,
  handleBarClick,
}) => {
  return (
    <div className={styles.base}>
      <Bar
        data={chartData}
        options={options}
        ref={chartRef}
        style={{height: pipelinesLength * Y_AXIS_HEIGHT}}
        onClick={handleBarClick}
        aria-label="Runtimes Chart"
      >
        Your browser does not support the canvas element
      </Bar>
    </div>
  );
};

export default BarChart;
