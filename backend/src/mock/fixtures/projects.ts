import {Project, ProjectInfo} from '@dash-backend/proto';

const projects: Record<string, ProjectInfo> = {
  'Solar-Panel-Data-Sorting': new ProjectInfo()
    .setProject(new Project().setName('Solar-Panel-Data-Sorting'))
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
  'Data-Cleaning-Process': new ProjectInfo()
    .setProject(new Project().setName('Data-Cleaning-Process'))
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
  'Solar-Power-Data-Logger-Team-Collab': new ProjectInfo()
    .setProject(new Project().setName('Solar-Power-Data-Logger-Team-Collab'))
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
  'Solar-Price-Prediction-Modal': new ProjectInfo()
    .setProject(new Project().setName('Solar-Price-Prediction-Modal'))
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
  'Egress-Examples': new ProjectInfo()
    .setProject(new Project().setName('Egress-Examples'))
    .setDescription(
      'Multiple pipelines outputting to different forms of egress',
    ),
  'Empty-Project': new ProjectInfo()
    .setProject(new Project().setName('Empty-Project'))
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
  'Trait-Discovery': new ProjectInfo()
    .setProject(new Project().setName('Trait-Discovery'))
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
  'Multi-Project-Pipeline-A': new ProjectInfo()
    .setProject(new Project().setName('Multi-Project-Pipeline-A'))
    .setDescription(
      'Contains two DAGs spanning across this and Multi-Project-Pipeline-B',
    ),
  'Multi-Project-Pipeline-B': new ProjectInfo()
    .setProject(new Project().setName('Multi-Project-Pipeline-B'))
    .setDescription(
      'Contains two DAGs spanning across this and Multi-Project-Pipeline-A',
    ),
};

const tutorialProjects: {[projectId: string]: ProjectInfo} = {
  'OpenCV-Tutorial': new ProjectInfo()
    .setProject(new Project().setName('OpenCV-Tutorial'))
    .setDescription(
      'Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    ),
};

const loadProjects: {[projectId: string]: ProjectInfo} = {
  'Load-Project': new ProjectInfo()
    .setProject(new Project().setName('Load-Project'))
    .setDescription('Project for testing frontend load'),
};

export const projectInfo = projects;

export const allProjects = {...projects, ...tutorialProjects, ...loadProjects};

export default projects;
