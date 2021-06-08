import capitalize from 'lodash/capitalize';

import {JobState} from '@graphqlTypes';

const readableJobState = (jobState: JobState | string) => {
  const state = jobState.toString().replace('JOB_', '');
  return capitalize(state);
};

export default readableJobState;
