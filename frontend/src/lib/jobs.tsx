import capitalize from 'lodash/capitalize';
import React from 'react';

import {JobState} from '@dash-frontend/api/pps';
import {
  formatDurationFromSeconds,
  formatDurationFromSecondsToNow,
} from '@dash-frontend/lib/dateTime';
import {
  StatusWarningSVG,
  StatusDotsSVG,
  StatusCheckmarkSVG,
  Icon,
} from '@pachyderm/components';

type JobVisualState = 'RUNNING' | 'ERROR' | 'SUCCESS' | 'IDLE';

export const getJobStateIcon = (state: JobVisualState, small?: boolean) => {
  switch (state) {
    case 'ERROR':
      return (
        <Icon small={small} color="red">
          <StatusWarningSVG />
        </Icon>
      );
    case 'RUNNING':
      return (
        <Icon small={small} color="green">
          <StatusDotsSVG />
        </Icon>
      );
    case 'SUCCESS':
      return (
        <Icon small={small} color="green">
          <StatusCheckmarkSVG />
        </Icon>
      );
    default:
      return null;
  }
};

export const getJobStateSVG = (state: JobVisualState) => {
  switch (state) {
    case 'ERROR':
      return StatusWarningSVG;
    case 'RUNNING':
      return StatusDotsSVG;
    case 'SUCCESS':
      return StatusCheckmarkSVG;
    default:
      return null;
  }
};

export const getJobStateColor = (state: JobVisualState) => {
  switch (state) {
    case 'ERROR':
      return 'red';
    case 'RUNNING':
    case 'SUCCESS':
      return 'green';
    default:
      return null;
  }
};

export const readableJobState = (jobState: JobState | string) => {
  const state = jobState.toString().replace(/JOB_(STATE_)?/, '');
  return capitalize(state);
};

export const getVisualJobState = (state?: JobState): JobVisualState => {
  switch (state) {
    case JobState.JOB_SUCCESS:
    case JobState.JOB_CREATED:
      return 'SUCCESS';
    case JobState.JOB_EGRESSING:
    case JobState.JOB_RUNNING:
    case JobState.JOB_STARTING:
    case JobState.JOB_FINISHING:
      return 'RUNNING';
    case JobState.JOB_FAILURE:
    case JobState.JOB_KILLED:
    case JobState.JOB_UNRUNNABLE:
      return 'ERROR';
    default:
      return 'IDLE';
  }
};

export const getJobRuntime = (
  startedAt?: number | null,
  finishedAt?: number | null,
) => {
  if (finishedAt && startedAt) {
    return formatDurationFromSeconds(finishedAt - startedAt);
  }
  if (startedAt) {
    return `${formatDurationFromSecondsToNow(startedAt)} - In Progress`;
  }
  return 'In Progress';
};
