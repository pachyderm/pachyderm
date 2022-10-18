import {JobState} from '@graphqlTypes';
import {
  StatusWarningSVG,
  StatusDotsSVG,
  StatusCheckmarkSVG,
  Icon,
} from '@pachyderm/components';
import capitalize from 'lodash/capitalize';
import React from 'react';

type JobVisualState = 'RUNNING' | 'ERROR' | 'SUCCESS' | 'IDLE';

export const getJobStateIcon = (state: JobVisualState) => {
  switch (state) {
    case 'ERROR':
      return (
        <Icon color="red">
          <StatusWarningSVG />
        </Icon>
      );
    case 'RUNNING':
      return (
        <Icon color="green">
          <StatusDotsSVG />
        </Icon>
      );
    case 'SUCCESS':
      return (
        <Icon color="green">
          <StatusCheckmarkSVG />
        </Icon>
      );
    default:
      return null;
  }
};

export const readableJobState = (jobState: JobState | string) => {
  const state = jobState.toString().replace(/JOB_(STATE_)?/, '');
  return capitalize(state);
};

// should match job state mappings from backend/src/lib/nodeStateMappers
export const getVisualJobState = (state: JobState): JobVisualState => {
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
