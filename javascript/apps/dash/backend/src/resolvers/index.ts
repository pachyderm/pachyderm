import merge from 'lodash/merge';

import {Resolvers} from 'generated/types';

import dagResolver from './Dag';
import repoResolver from './Repo';

const resolver: Resolvers = merge(dagResolver, repoResolver, {});

export default resolver;
