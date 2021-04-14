import merge from 'lodash/merge';

import authenticated from '@dash-backend/middleware/authenticated';
import {
  PipelineState,
  Resolvers,
  JobState,
  ProjectStatus,
  FileType,
} from '@graphqlTypes';

import authResolver from './Auth';
import dagResolver from './Dag';
import fileResolver from './File';
import jobResolver from './Job';
import pipelineResolver from './Pipeline';
import projectsResolver from './Projects';
import repoResolver from './Repo';
import searchResolver from './Search';

const resolvers: Resolvers = merge(
  {JobState: JobState},
  {PipelineState: PipelineState},
  {ProjectStatus: ProjectStatus},
  {FileType: FileType},
  fileResolver,
  dagResolver,
  repoResolver,
  authResolver,
  projectsResolver,
  jobResolver,
  searchResolver,
  pipelineResolver,
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
