import {ChartData} from 'chart.js';
import {parseISO} from 'date-fns';
import {draw} from 'patternomaly';
import {useMemo, useCallback} from 'react';

import {JobInfo} from '@dash-frontend/api/pps';

import {getChartColor} from '../RuntimesChart';

const SECONDS_IN_HOUR = 60 * 60;

const useRuntimesChartData = (filteredJobs: JobInfo[], selectedJob: string) => {
  const jobIds = useMemo(
    () => [...new Set(filteredJobs.map((job) => job?.job?.id || ''))],
    [filteredJobs],
  );
  const pipelines = useMemo(
    () => [
      ...new Set(filteredJobs.map((job) => `@${job?.job?.pipeline?.name}`)),
    ],
    [filteredJobs],
  );

  // create a reference of jobs where job = jobsCrossReference[jobId][pipeline]
  const jobsCrossReference = useMemo(() => {
    const jobsCrossReference: Record<string, Record<string, JobInfo>> = {};
    filteredJobs.forEach((job) => {
      if (!jobsCrossReference[job?.job?.id || '']) {
        jobsCrossReference[job?.job?.id || ''] = {};
      }
      jobsCrossReference[job?.job?.id || ''][`@${job?.job?.pipeline?.name}`] =
        job;
    });
    return jobsCrossReference;
  }, [filteredJobs]);

  const getJobDuration = useCallback(
    (id: string, step: string) => {
      const job = jobsCrossReference[id][step];
      if (job && job.started) {
        return (
          (job.finished
            ? parseISO(job.finished).getTime() / 1000
            : Math.floor(Date.now() / 1000)) -
          (job.started ? parseISO(job.started).getTime() / 1000 : 0)
        );
      } else {
        return null;
      }
    },
    [jobsCrossReference],
  );

  const longestJob = useMemo(
    () =>
      Math.max(
        ...filteredJobs.map(
          (job) =>
            getJobDuration(
              job?.job?.id || '',
              `@${job?.job?.pipeline?.name}`,
            ) || 0,
        ),
      ),
    [filteredJobs, getJobDuration],
  );
  const useHoursAsUnit = longestJob > SECONDS_IN_HOUR;

  const pipelineDatasets = useMemo(() => {
    return jobIds.map((id, index) => {
      // create an array of [a: number, b: number] for each pipeline job in a
      // jobSet where a is the arbitrary point in time the job started, starting
      // at 0 and b is that point in time plus the job's duration.
      // ex: [[0,15],[15,30]] for 2 jobs in a DAG that took 15 seconds each.
      const jobSetData: ([number, number] | null)[] = [];
      let durationSoFar = 0;
      const backgroundColors: (string | CanvasPattern)[] = [];
      const colorOpacity = selectedJob && selectedJob !== id ? 0.25 : 1;

      pipelines.forEach((pipeline) => {
        const latestDuration = getJobDuration(id, pipeline);
        const outputDuration =
          latestDuration !== null
            ? latestDuration / (useHoursAsUnit ? SECONDS_IN_HOUR : 1)
            : null;
        jobSetData.push(
          outputDuration
            ? [durationSoFar, durationSoFar + outputDuration]
            : null,
        );
        durationSoFar += outputDuration || 0;

        if (
          jobsCrossReference[id][pipeline] &&
          Number(jobsCrossReference[id][pipeline]?.dataFailed) > 0
        ) {
          backgroundColors.push(
            draw('zigzag', getChartColor(index, colorOpacity)),
          );
        } else if (
          jobsCrossReference[id][pipeline] &&
          !jobsCrossReference[id][pipeline].finished
        ) {
          backgroundColors.push(
            draw('diagonal', getChartColor(index, colorOpacity)),
          );
        } else {
          backgroundColors.push(getChartColor(index, colorOpacity));
        }
      });

      return {
        label: id,
        data: jobSetData,
        minBarLength: 8,
        backgroundColor: backgroundColors,
      };
    });
  }, [
    jobIds,
    pipelines,
    selectedJob,
    getJobDuration,
    useHoursAsUnit,
    jobsCrossReference,
  ]);

  const chartData: ChartData<'bar'> = {
    labels: pipelines,
    datasets: pipelineDatasets,
  };

  const jobsWithFailedDatums = useMemo(
    () => filteredJobs?.filter((job) => Number(job?.dataFailed) > 0),
    [filteredJobs],
  );

  return {
    chartData,
    pipelineDatasets,
    jobsCrossReference,
    pipelines,
    jobIds,
    useHoursAsUnit,
    jobsWithFailedDatums,
    longestJob,
  };
};

export default useRuntimesChartData;
