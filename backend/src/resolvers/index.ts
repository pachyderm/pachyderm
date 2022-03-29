import merge from 'lodash/merge';

import authenticated from '@dash-backend/middleware/authenticated';
import {Resolvers} from '@graphqlTypes';

import authResolver from './Auth';
import branchResolver from './Branch';
import commitResolver from './Commit';
import dagResolver from './Dag';
import datumResolver from './Datum';
import fileResolver from './File';
import jobResolver from './Job';
import logsResolver from './Logs';
import pipelineResolver from './Pipeline';
import projectsResolver from './Projects';
import repoResolver from './Repo';
import searchResolver from './Search';

const resolvers: Resolvers = merge(
  fileResolver,
  dagResolver,
  datumResolver,
  logsResolver,
  repoResolver,
  authResolver,
  projectsResolver,
  jobResolver,
  searchResolver,
  pipelineResolver,
  commitResolver,
  branchResolver,
  {},
);

// NOTE: This does not support field level resolvers
const unauthenticated = ['exchangeCode', 'authConfig'];

Object.keys(resolvers.Query || {}).forEach((resolver) => {
  if (!unauthenticated.includes(resolver) && resolvers.Query) {
    resolvers.Query[resolver] = authenticated(resolvers.Query[resolver]);
  }
});

Object.keys(resolvers.Mutation || {}).forEach((resolver) => {
  if (!unauthenticated.includes(resolver) && resolvers.Mutation) {
    resolvers.Mutation[resolver] = authenticated(resolvers.Mutation[resolver]);
  }
});

export default resolvers;
