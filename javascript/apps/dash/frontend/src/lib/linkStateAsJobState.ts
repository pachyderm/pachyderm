import {Link, JobState} from '@graphqlTypes';

const linkStateAsJobState = (state: Link['state']) => {
  return JobState[(state as unknown) as keyof typeof JobState];
};

export default linkStateAsJobState;
