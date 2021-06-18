import {
  Project,
  Projects,
  ProjectStatus,
} from '@pachyderm/proto/pb/projects/projects_pb';
import {Timestamp} from 'google-protobuf/google/protobuf/timestamp_pb';

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
    .setName('Solar Industry Analysis 2020')
    .setCreatedat(new Timestamp().setSeconds(1614126189))
    .setStatus(ProjectStatus.UNHEALTHY)
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
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

export const projectInfo = new Projects().setProjectInfoList(
  Object.values(projects),
);

export default projects;
