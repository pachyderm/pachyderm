import {render, screen} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import {useProjects} from '../useProjects';

const ProjectsComponent = withContextProviders(() => {
  const {projects, loading} = useProjects();

  if (loading) return <span>Loading</span>;

  return (
    <div>
      {(projects || []).map((project, i) => (
        <div key={project.id}>
          <span>
            {i} project id: {project.id}
          </span>
          <span>
            {i} project name: {project.name}
          </span>
          <span>
            {i} project description: {project.description}
          </span>
          <span>
            {i} project createdAt: {project.createdAt}
          </span>
          <span>
            {i} project status: {project.status}
          </span>
        </div>
      ))}
    </div>
  );
});

describe('useProjects', () => {
  it('should get projects', async () => {
    render(<ProjectsComponent />);
    const project0Id = await screen.findByText('0 project id: 1');
    const project1Id = await screen.findByText('1 project id: 2');
    const project0Name = await screen.findByText(
      '0 project name: Solar Panel Data Sorting',
    );
    const projectDescription = await screen.findByText(
      '0 project description: Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    );
    const projectCreationDate = await screen.findByText(
      '0 project createdAt: 1614026189',
    );
    const projectStatus = await screen.findByText('0 project status: HEALTHY');

    expect(project0Id).toBeInTheDocument();
    expect(project1Id).toBeInTheDocument();
    expect(project0Name).toBeInTheDocument();
    expect(projectDescription).toBeInTheDocument();
    expect(projectCreationDate).toBeInTheDocument();
    expect(projectStatus).toBeInTheDocument();
  });
});
