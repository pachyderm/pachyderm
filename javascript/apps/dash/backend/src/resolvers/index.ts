import merge from 'lodash/merge';

import {PipelineState, Resolvers} from 'generated/types';

import dagResolver from './Dag';
import repoResolver from './Repo';

const resolver: Resolvers = merge(
  {PipelineState: PipelineState},
  dagResolver,
  repoResolver,
  {},
);
export default resolver;
