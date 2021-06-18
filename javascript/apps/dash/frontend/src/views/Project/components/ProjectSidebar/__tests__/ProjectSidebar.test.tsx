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

      expect(queryByTestId('JobListStatic__loadingdots')).toBeInTheDocument();
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
      window.history.replaceState('', '', '/project/3/repo/cron/branch/master');

      const {findByTestId, getByText} = render(<Project />);

      const repoName = await findByTestId('Title__name');
      const size = getByText('607.28 KB');
      const commit = getByText('ID 9d5daa0918ac4c43a476b86e3bb5e88e');

      expect(repoName).toHaveTextContent('cron');
      expect(size).toBeInTheDocument();
      expect(commit).toBeInTheDocument();
    });

    it('should show no commits when the branch has no commits', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/2/repo/samples/branch/master',
      );

      const {findByText} = render(<Project />);

      const emptyMessage = await findByText(
        'There are no commits for this branch',
      );

      expect(emptyMessage).toBeInTheDocument();
    });

    it('should show empty repo message when repo has no commits', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/3/repo/processor/branch/master',
      );

      const {findByText} = render(<Project />);

      const emptyMessage = await findByText(
        'Commit your first file on this repo!',
      );

      expect(emptyMessage).toBeInTheDocument();
    });
  });
});
