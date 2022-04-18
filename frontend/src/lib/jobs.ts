import {JobState} from '@graphqlTypes';
import capitalize from 'lodash/capitalize';

type JobVisualState = 'BUSY' | 'ERROR' | 'SUCCESS';

export const getJobStateHref = (state: JobVisualState) => {
  switch (state) {
    case 'BUSY':
      return '/dag_busy.svg';
    case 'ERROR':
      return '/dag_error.svg';
    default:
      return '/dag_success.svg';
  }
};

export const readableJobState = (jobState: JobState | string) => {
  const state = jobState.toString().replace(/JOB_(STATE_)?/, '');
  return capitalize(state);
};

export const getVisualJobState = (state: JobState): JobVisualState => {
  switch (state) {
    case JobState.JOB_CREATED:
    case JobState.JOB_EGRESSING:
    case JobState.JOB_RUNNING:
    case JobState.JOB_STARTING:
    case JobState.JOB_FINISHING:
      return 'BUSY';
    case JobState.JOB_FAILURE:
    case JobState.JOB_KILLED:
      return 'ERROR';
    default:
      return 'SUCCESS';
  }
};
