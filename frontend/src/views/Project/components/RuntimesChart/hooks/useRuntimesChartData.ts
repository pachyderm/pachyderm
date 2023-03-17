import {Job} from '@graphqlTypes';
import {ChartData} from 'chart.js';
import {useMemo, useCallback} from 'react';

import {CHART_COLORS} from '../RuntimesChart';

const useRuntimesChartData = (filteredJobs: Job[]) => {
  const jobIds = useMemo(
    () => [...new Set(filteredJobs.map((job: Job) => job.id))],
    [filteredJobs],
  );
  const pipelines = useMemo(
    () => [...new Set(filteredJobs.map((job: Job) => `@${job.pipelineName}`))],
    [filteredJobs],
  );

  // create a reference of jobs where job = jobsCrossReference[jobId][pipeline]
  const jobsCrossReference = useMemo(() => {
    const jobsCrossReference: Record<string, Record<string, Job>> = {};
    filteredJobs.forEach((job: Job) => {
      if (!jobsCrossReference[job.id]) {
        jobsCrossReference[job.id] = {};
      }
      jobsCrossReference[job.id][`@${job.pipelineName}`] = job;
    });
    return jobsCrossReference;
  }, [filteredJobs]);

  const getJobDuration = useCallback(
    (id: string, step: string) => {
      const job = jobsCrossReference[id][step];
      if (job) {
        return (job.finishedAt || Date.now()) - (job.createdAt || 0);
      } else {
        return null;
      }
    },
    [jobsCrossReference],
  );

  const pipelineDatasets = useMemo(() => {
    return jobIds.map((id, index) => {
      // create an array of [a: number, b: number] for each pipeline job in a
      // jobSet where a is the arbitrary point in time the job started, starting
      // at 0 and b is that point in time plus the job's duration.
      // ex: [[0,15],[15,30]] for 2 jobs in a DAG that took 15 seconds each.
      const jobSetData: [number, number][] = [];
      let durationSoFar = 0;
      pipelines.forEach((pipeline) => {
        const latestDuration = getJobDuration(id, pipeline) || 0;
        jobSetData.push([durationSoFar, durationSoFar + latestDuration]);
        durationSoFar += latestDuration;
      });
      return {
        label: id,
        data: jobSetData,
        backgroundColor: CHART_COLORS[index % CHART_COLORS.length],
      };
    });
  }, [getJobDuration, jobIds, pipelines]);

  const chartData: ChartData<'bar'> = {
    labels: pipelines,
    datasets: pipelineDatasets,
  };

  return {chartData, pipelineDatasets, jobsCrossReference, pipelines, jobIds};
};

export default useRuntimesChartData;
