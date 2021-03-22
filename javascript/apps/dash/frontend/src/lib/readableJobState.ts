import capitalize from 'lodash/capitalize';

import {JobState} from '@graphqlTypes';

const readableJobState = (JobState: JobState) => {
  const state = JobState.toString().split('JOB_')[1];
  return capitalize(state);
};

export default readableJobState;
