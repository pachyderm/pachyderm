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
            {i} project name: {project.id}
          </span>
          <span>
            {i} project description: {project.description}
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
    const projectName = await screen.findByText(
      '0 project name: Solar-Panel-Data-Sorting',
    );
    const projectDescription = await screen.findByText(
      '0 project description: Lorem ipsum dolor sit amet, consectetu adipiscing elit, sed do eiusmod tempor',
    );
    const projectStatus = await screen.findByText(
      '0 project status: UNHEALTHY',
    );

    expect(projectName).toBeInTheDocument();
    expect(projectDescription).toBeInTheDocument();
    expect(projectStatus).toBeInTheDocument();
  });
});
