import {Timestamp} from 'google-protobuf/google/protobuf/timestamp_pb';

import {Project, Projects, ProjectStatus} from '@dash-backend/proto';

const projects: {[projectId: string]: Project} = {
  '1': new Project()
    .setId('1')
    .setName('Solar Panel Data Sorting')
    .setCreatedat(new Timestamp().setSeconds(1614026189))
    .setStatus(ProjectStatus.HEALTHY)
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
  '2': new Project()
    .setId('2')
    .setName('Data Cleaning Process')
    .setCreatedat(new Timestamp().setSeconds(1614526189))
    .setStatus(ProjectStatus.UNHEALTHY)
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
  '3': new Project()
    .setId('3')
    .setName('Solar Power Data Logger Team Collab')
    .setCreatedat(new Timestamp().setSeconds(1614126189))
    .setStatus(ProjectStatus.HEALTHY)
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
  '4': new Project()
    .setId('4')
    .setName('Solar Price Prediction Modal')
    .setCreatedat(new Timestamp().setSeconds(1614126189))
    .setStatus(ProjectStatus.HEALTHY)
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
  '5': new Project()
    .setId('5')
    .setName('Egress Examples')
    .setCreatedat(new Timestamp().setSeconds(1614126189))
    .setStatus(ProjectStatus.UNHEALTHY)
    .setDescription(
      'Multiple pipelines outputting to different forms of egress',
    ),
  '6': new Project()
    .setId('6')
    .setName('Empty Project')
    .setCreatedat(new Timestamp().setSeconds(1614126189))
    .setStatus(ProjectStatus.HEALTHY)
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
  '7': new Project()
    .setId('7')
    .setName('Trait Discovery')
    .setCreatedat(new Timestamp().setSeconds(1614126189))
    .setStatus(ProjectStatus.HEALTHY)
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
};

const tutorialProjects: {[projectId: string]: Project} = {
  '8': new Project()
    .setId('8')
    .setName('OpenCV Tutorial')
    .setCreatedat(new Timestamp().setSeconds(1614126189))
    .setStatus(ProjectStatus.HEALTHY)
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
};

const loadProjects: {[projectId: string]: Project} = {
  '9': new Project()
    .setId('9')
    .setName('Load Project')
    .setCreatedat(new Timestamp().setSeconds(1651264871))
    .setStatus(ProjectStatus.HEALTHY)
    .setDescription('Project for testing frontend load'),
};

export const projectInfo = new Projects().setProjectInfoList(
  Object.values(projects),
);

export const allProjects = {...projects, ...tutorialProjects, ...loadProjects};

export default projects;
