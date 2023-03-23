import {ApolloError} from '@apollo/client';
import {JobsQuery, Job} from '@graphqlTypes';
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
import React, {useRef, useCallback, useMemo} from 'react';
import {getElementAtEvent} from 'react-chartjs-2';
import {useHistory} from 'react-router-dom';

import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import {
  TableViewFilters,
  TableViewLoadingDots,
} from '@dash-frontend/components/TableView';
import {MAX_FILTER_HEIGHT_REM} from '@dash-frontend/components/TableView/components/TableViewFilters/TableViewFilters';
import useLogsNavigation from '@dash-frontend/hooks/useLogsNavigation';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {Form} from '@pachyderm/components';

import BarChart from './components/BarChart';
import Tooltip from './components/Tooltip';
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
  jobs?: JobsQuery['jobs'];
  loading: boolean;
  error?: ApolloError;
  resource: 'job' | 'pipeline';
  viewOptions?: JSX.Element;
  filtersExpanded: boolean;
};

export const CHART_COLORS = [
  '#6929c4',
  '#1192e8',
  '#005d5d',
  '#9f1853',
  '#fa4d56',
  '#570408',
  '#198038',
  '#002d9c',
  '#ee538b',
  '#b28600',
  '#009d9a',
  '#012749',
  '#8a3800',
  '#a56eff',
];

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
  } = useRuntimesChartData(filteredJobs);
  const {tooltip, setTooltipState} = useRuntimesChartTooltip();

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
            x: {min: 0},
          },
          pan: {
            enabled: true,
            mode: 'x',
          },
          zoom: {
            wheel: {
              enabled: true,
            },
            mode: 'x',
          },
        },
      },
    }),
    [setTooltipState, useHoursAsUnit],
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
            {viewOptions && (
              <div className={styles.viewOptions}>{viewOptions}</div>
            )}
            <div className={styles.legend}>
              {Object.keys(jobsCrossReference).map((jobId, index) => {
                const oldestJob = Math.min(
                  ...Object.values(jobsCrossReference[jobId]).map(
                    (job: Job) => job.createdAt || 0,
                  ),
                );
                return (
                  <div key={jobId} className={styles.legendItem}>
                    <div
                      className={styles.legendBox}
                      style={{
                        backgroundColor:
                          CHART_COLORS[index % CHART_COLORS.length],
                      }}
                    />
                    {oldestJob && getStandardDate(oldestJob)};{' '}
                    {jobId.slice(0, 6)}
                    ...
                  </div>
                );
              })}
              {/* To be added with datum errors */}
              {/* <div className={styles.legendItem}>
              <div className={styles.failedBox} />
              Failed Datums
            </div> */}
            </div>
          </div>
          <BarChart
            chartData={chartData}
            options={options}
            chartRef={chartRef}
            pipelinesLength={pipelines.length}
            handleBarClick={handleClick}
          />
          <Tooltip tooltipState={tooltip} useHoursAsUnit={useHoursAsUnit} />
        </div>
      )}
    </Form>
  );
};

export default RuntimesChart;
