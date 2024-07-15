import {
  Chart as ChartJS,
  ChartOptions,
  BarElement,
  Title,
  Legend,
  LinearScale,
  CategoryScale,
} from 'chart.js';
import zoomPlugin from 'chartjs-plugin-zoom';
import React, {useRef, useCallback, useMemo, useState} from 'react';
import {getElementAtEvent} from 'react-chartjs-2';
import {useHistory} from 'react-router-dom';

import {JobInfo} from '@dash-frontend/api/pps';
import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import {
  TableViewFilters,
  TableViewLoadingDots,
} from '@dash-frontend/components/TableView';
import {MAX_FILTER_HEIGHT_REM} from '@dash-frontend/components/TableView/components/TableViewFilters/TableViewFilters';
import useLogsNavigation from '@dash-frontend/hooks/useLogsNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  getStandardDateFromUnixSeconds,
  getUnixSecondsFromISOString,
} from '@dash-frontend/lib/dateTime';
import {Form, Button} from '@pachyderm/components';

import BarChart from './components/BarChart';
import DatumsSidebar from './components/DatumsSidebar';
import Tooltip from './components/Tooltip';
import ZoomPanSlider from './components/ZoomPanSlider';
import useRuntimesChartData from './hooks/useRuntimesChartData';
import useRuntimesChartFilters from './hooks/useRuntimesChartFilters';
import useRuntimesChartTooltip from './hooks/useRuntimesChartTooltip';
import styles from './RuntimesChart.module.css';

const TABLEVIEW_HEADER_OFFSET = 13.5 * 16;

ChartJS.register(
  BarElement,
  Title,
  Legend,
  CategoryScale,
  LinearScale,
  zoomPlugin,
);

type RuntimesChartProps = {
  jobs?: JobInfo[];
  loading: boolean;
  error?: string;
  resource: 'job' | 'pipeline';
  viewOptions?: JSX.Element;
  filtersExpanded: boolean;
};

export const getChartColor = (index: number, opacity = 1) => {
  const chartColorsRGB = [
    [105, 41, 196],
    [17, 146, 232],
    [0, 93, 93],
    [159, 23, 83],
    [250, 77, 86],
    [87, 4, 8],
    [25, 128, 56],
    [0, 45, 156],
    [238, 83, 139],
    [178, 134, 0],
    [0, 157, 154],
    [1, 39, 73],
    [138, 56, 0],
    [165, 110, 255],
  ];
  const [r, g, b] = chartColorsRGB[index % chartColorsRGB.length];
  return `rgba(${r},${g},${b},${opacity})`;
};

const DEFAULT_FONT = {
  family: 'Public Sans',
  weight: '300',
  size: 14,
};

const RuntimesChart: React.FC<RuntimesChartProps> = ({
  jobs,
  loading,
  error,
  resource,
  viewOptions,
  filtersExpanded,
}) => {
  const chartRef = useRef<ChartJS<'bar'>>(null);
  const {projectId} = useUrlState();
  const browserHistory = useHistory();
  const [datumsSidebarOpen, setDatumsSidebarOpen] = useState(false);
  const [selectedJob, selectFailedDatumJob] = useState('');
  const {getPathToDatumLogs} = useLogsNavigation();
  const {filteredJobs, formCtx, clearableFiltersMap, multiselectFilters} =
    useRuntimesChartFilters({jobs});
  const {
    chartData,
    pipelineDatasets,
    jobsCrossReference,
    pipelines,
    jobIds,
    useHoursAsUnit,
    jobsWithFailedDatums,
    longestJob,
  } = useRuntimesChartData(filteredJobs, selectedJob);
  const {tooltip, setTooltipState} =
    useRuntimesChartTooltip(jobsCrossReference);

  const options: ChartOptions<'bar'> = useMemo(
    () => ({
      animation: false,
      indexAxis: 'y',
      layout: {
        padding: 20,
      },
      scales: {
        y: {
          ticks: {
            font: DEFAULT_FONT,
          },
          beginAtZero: true,
          border: {
            color: '#000',
          },
        },
        x: {
          ticks: {
            font: DEFAULT_FONT,
          },
          title: {
            display: true,
            text: `Job Duration (${useHoursAsUnit ? 'Hours' : 'Seconds'})`,
            font: {
              family: 'Public Sans',
              weight: '600',
              size: 12,
            },
          },
          beginAtZero: true,
          position: 'top',
          border: {
            color: '#000',
          },
        },
      },
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          display: false,
        },
        tooltip: {
          enabled: false,
          external: setTooltipState,
        },
        zoom: {
          limits: {
            x: {min: 0, max: longestJob},
          },
        },
      },
    }),
    [longestJob, setTooltipState, useHoursAsUnit],
  );

  const handleClick = useCallback(
    (event: React.MouseEvent<HTMLCanvasElement, MouseEvent>) => {
      const element = chartRef.current
        ? getElementAtEvent(chartRef.current, event)
        : [];
      if (element[0]) {
        const jobId = pipelineDatasets[element[0].datasetIndex].label;
        const pipelineId = pipelines[element[0].index].slice(1);
        const logsLink = getPathToDatumLogs(
          {
            projectId,
            jobId,
            pipelineId,
          },
          [],
        );
        browserHistory.push(logsLink);
      }
    },
    [
      browserHistory,
      getPathToDatumLogs,
      pipelineDatasets,
      pipelines,
      projectId,
    ],
  );

  if (loading) {
    return <TableViewLoadingDots data-testid="RuntimesChart__loadingDots" />;
  }

  if (error) {
    return (
      <ErrorStateSupportLink
        title="We couldn't load the jobs list"
        message="Your jobs have been processed, but we couldn't fetch a list of them from our end. Please try refreshing this page."
      />
    );
  }

  return (
    <Form formContext={formCtx}>
      <TableViewFilters
        formCtx={formCtx}
        filtersExpanded={filtersExpanded}
        multiselectFilters={multiselectFilters}
        clearableFiltersMap={clearableFiltersMap}
      />
      {filteredJobs?.length === 0 ? (
        <EmptyState
          title="No matching results"
          message="We couldn't find any results matching your filters. Try using a different set of filters."
        />
      ) : (
        <div
          className={styles.base}
          style={{
            height: `calc(100vh - ${TABLEVIEW_HEADER_OFFSET}px - ${
              filtersExpanded ? MAX_FILTER_HEIGHT_REM * 16 : 0
            }px`,
          }}
          data-testid="RuntimesChart__chart"
        >
          <div className={styles.titleContent}>
            <div className={styles.titleLine}>
              {resource === 'job' && (
                <h5>
                  Runtimes for {jobIds.length}{' '}
                  {jobIds.length > 1 ? 'jobs' : 'job'}
                </h5>
              )}
              {resource === 'pipeline' && (
                <h5>
                  Runtimes for {pipelines.length}{' '}
                  {pipelines.length > 1 ? 'pipelines' : 'pipeline'}
                </h5>
              )}
              {jobsWithFailedDatums.length > 0 && (
                <Button
                  buttonType="secondary"
                  onClick={() => setDatumsSidebarOpen(!datumsSidebarOpen)}
                >
                  Jobs with failed datums
                </Button>
              )}
            </div>
            {viewOptions && (
              <div className={styles.viewOptions}>{viewOptions}</div>
            )}
            <div className={styles.legend}>
              {Object.keys(jobsCrossReference).map((jobId, index) => {
                const oldestJob = Math.min(
                  ...Object.values(jobsCrossReference[jobId]).map((job) =>
                    getUnixSecondsFromISOString(job.created),
                  ),
                );
                return (
                  <div key={jobId} className={styles.legendItem}>
                    <div
                      className={styles.legendBox}
                      style={{
                        backgroundColor: getChartColor(index),
                      }}
                    />
                    {oldestJob && getStandardDateFromUnixSeconds(oldestJob)};{' '}
                    {jobId.slice(0, 6)}
                    ...
                  </div>
                );
              })}
              <div className={styles.halfLegendItem}>
                <div className={styles.failedBox} />
                Failed Datums
              </div>
              <div className={styles.halfLegendItem}>
                <div className={styles.inProgressBox} />
                In Progress
              </div>
            </div>
          </div>
          <ZoomPanSlider chartRef={chartRef} />
          <BarChart
            chartData={chartData}
            options={options}
            chartRef={chartRef}
            pipelinesLength={pipelines.length}
            maxJobsPerPipeline={pipelineDatasets.length}
            handleBarClick={handleClick}
          />
          <Tooltip tooltipState={tooltip} useHoursAsUnit={useHoursAsUnit} />
          {datumsSidebarOpen && (
            <DatumsSidebar
              jobs={jobsWithFailedDatums}
              selectedJob={selectedJob}
              selectFailedDatumJob={selectFailedDatumJob}
              onClose={() => setDatumsSidebarOpen(false)}
            />
          )}
        </div>
      )}
    </Form>
  );
};

export default RuntimesChart;
