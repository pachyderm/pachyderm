import merge from 'lodash/merge';

import authenticated from '@dash-backend/middleware/authenticated';
import {Resolvers} from '@graphqlTypes';

import adminResolver from './Admin';
import authResolver from './Auth';
import branchResolver from './Branch';
import commitResolver from './Commit';
import datumResolver from './Datum';
import enterpriseResolver from './Enterprise';
import fileResolver from './File';
import jobResolver from './Job';
import logsResolver from './Logs';
import pipelineResolver from './Pipeline';
import projectsResolver from './Projects';
import repoResolver from './Repo';
import searchResolver from './Search';
import versionResolver from './Version';
import verticesResolver from './Vertices';

const resolvers: Resolvers = merge(
  fileResolver,
  verticesResolver,
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
  enterpriseResolver,
  adminResolver,
  versionResolver,
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
