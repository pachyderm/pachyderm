import capitalize from 'lodash/capitalize';

import {JobState} from '@graphqlTypes';

const readableJobState = (JobState: JobState | string) => {
  const state = JobState.toString().replace('JOB_', '');
  return capitalize(state);
};

export default readableJobState;
