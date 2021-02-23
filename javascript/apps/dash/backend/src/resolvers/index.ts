import merge from 'lodash/merge';

import {PipelineState, Resolvers, JobState} from 'generated/types';

import authResolver from './Auth';
import dagResolver from './Dag';
import repoResolver from './Repo';

const resolver: Resolvers = merge(
  {JobState: JobState},
  {PipelineState: PipelineState},
  dagResolver,
  repoResolver,
  authResolver,
  {},
);
export default resolver;
