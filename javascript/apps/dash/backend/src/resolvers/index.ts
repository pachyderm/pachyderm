import merge from 'lodash/merge';

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
import projectsResolver from './Projects';
import repoResolver from './Repo';

const resolver: Resolvers = merge(
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
  {},
);
export default resolver;
