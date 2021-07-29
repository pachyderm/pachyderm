import {JobState} from '@graphqlTypes';
import capitalize from 'lodash/capitalize';

const readableJobState = (jobState: JobState | string) => {
  const state = jobState.toString().replace(/JOB_(STATE_)?/, '');
  return capitalize(state);
};

export default readableJobState;
