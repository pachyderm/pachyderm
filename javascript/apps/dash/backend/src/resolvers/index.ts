import merge from 'lodash/merge';

import {PipelineState, Resolvers, JobState} from 'generated/types';

import dagResolver from './Dag';
import repoResolver from './Repo';

const resolver: Resolvers = merge(
  {JobState: JobState},
  {PipelineState: PipelineState},
  dagResolver,
  repoResolver,
  {},
);
export default resolver;
