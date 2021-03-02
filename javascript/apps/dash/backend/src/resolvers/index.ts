import merge from 'lodash/merge';

import {
  PipelineState,
  Resolvers,
  JobState,
  ProjectStatus,
} from 'generated/types';

import authResolver from './Auth';
import dagResolver from './Dag';
import projectsResolver from './Projects';
import repoResolver from './Repo';

const resolver: Resolvers = merge(
  {JobState: JobState},
  {PipelineState: PipelineState},
  {ProjectStatus: ProjectStatus},
  dagResolver,
  repoResolver,
  authResolver,
  projectsResolver,
  {},
);
export default resolver;
