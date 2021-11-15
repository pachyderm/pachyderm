import {render, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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

    it('should not display logs button', async () => {
      window.history.replaceState('', '', '/project/1/jobs');

      const {queryByText} = render(<Project />);
      expect(queryByText('Read Logs')).toBeNull();
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

    it('should display pipeline logs button', async () => {
      window.history.replaceState('', '', '/project/1/pipeline/montage');

      const {getByText} = render(<Project />);
      const logsLink = getByText('Read Logs').parentElement;
      expect(logsLink as HTMLElement).toHaveAttribute(
        'href',
        `/project/1/pipeline/montage/logs`,
      );
    });
  });

  describe('repos', () => {
    it('should display repo details', async () => {
      window.history.replaceState('', '', '/project/3/repo/cron/branch/master');

      const {findByTestId, getByText} = render(<Project />);

      const repoName = await findByTestId('Title__name');
      const size = getByText('607.28 KB');

      expect(repoName).toHaveTextContent('cron');
      expect(size).toBeInTheDocument();
      await waitFor(() =>
        expect(
          getByText('ID 9d5daa0918ac4c43a476b86e3bb5e88e'),
        ).toBeInTheDocument(),
      );
    });

    it('should not show a linked job when there is no job for the commit', async () => {
      window.history.replaceState('', '', '/project/3/repo/cron/branch/master');

      const {findByTestId, queryByRole} = render(<Project />);

      await findByTestId('Title__name');

      expect(queryByRole('link', {name: 'Linked Job'})).toBeNull();
    });

    it('should show a linked job for a commit', async () => {
      window.history.replaceState('', '', '/project/2/repo/test/branch/master');

      const {queryByRole} = render(<Project />);

      await waitFor(() =>
        expect(queryByRole('link', {name: 'Linked Job'})).toBeInTheDocument(),
      );
    });

    it('should show a linked job for a input repo commit', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/2/repo/training/branch/master',
      );

      const {queryByRole} = render(<Project />);

      await waitFor(() =>
        expect(queryByRole('link', {name: 'Linked Job'})).toBeInTheDocument(),
      );
    });

    it('should show no commits when the branch has no commits', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/2/repo/training/branch/develop',
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

    it('should not display logs button', async () => {
      window.history.replaceState('', '', '/project/3/repo/cron/branch/master');

      const {queryByText} = render(<Project />);
      expect(queryByText('Read Logs')).toBeNull();
    });
  });

  it('should filter commits by auto origin', async () => {
    window.history.replaceState('', '', '/project/3/repo/cron/branch/master');

    const {findByLabelText, queryAllByText} = render(<Project />);

    const hideAutoCommits = await findByLabelText('Show auto commits');
    expect(queryAllByText('View Files').length).toBe(3);
    userEvent.click(hideAutoCommits);
    await waitFor(() => expect(queryAllByText('View Files').length).toBe(2));
  });
});
