import {render} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import ProjectSidebar from '../ProjectSidebar';

describe('ProjectSidebar', () => {
  const Project = withContextProviders(ProjectSidebar);

  beforeEach(() => {
    window.history.replaceState('', '', '/');
  });

  it('should not display the sidebar if not on a sidebar route', async () => {
    window.history.replaceState('', '', '/project/1');

    const {queryByTestId} = render(<Project />);

    expect(queryByTestId('ProjectSidebar__sidebar')).not.toBeInTheDocument();
  });

  describe('jobs', () => {
    it('should display job list', async () => {
      window.history.replaceState('', '', '/project/1/jobs');

      const {queryByTestId, findByTestId} = render(<Project />);

      expect(queryByTestId('JobListSkeleton__list')).toBeInTheDocument();
      expect(await findByTestId('JobList__project1')).toBeInTheDocument();
    });
  });

  describe('pipelines', () => {
    it('should display pipeline details', async () => {
      window.history.replaceState('', '', '/project/1/pipeline/montage');

      const {queryByTestId, findByTestId} = render(<Project />);

      expect(
        queryByTestId('PipelineDetails__pipelineNameSkeleton'),
      ).toBeInTheDocument();

      const pipelineName = await findByTestId('Title__name');

      expect(pipelineName).toHaveTextContent('montage');
    });
  });

  describe('repos', () => {
    it('should display repo details', async () => {
      // TODO: update this with _actual_ repo details
      window.history.replaceState('', '', '/project/1/repo/montage');

      const {findByTestId} = render(<Project />);

      const repoName = await findByTestId('Title__name');

      expect(repoName).toHaveTextContent('montage');
    });
  });
});
