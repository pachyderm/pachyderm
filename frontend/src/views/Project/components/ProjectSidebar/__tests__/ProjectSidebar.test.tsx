import {
  render,
  waitFor,
  waitForElementToBeRemoved,
  screen,
} from '@testing-library/react';
import React from 'react';

import {
  withContextProviders,
  click,
  mockServer,
} from '@dash-frontend/testHelpers';

import ProjectSidebar from '../ProjectSidebar';

describe('ProjectSidebar', () => {
  const Project = withContextProviders(ProjectSidebar);

  beforeEach(() => {
    window.history.replaceState('', '', '/');
  });

  it('should not display the sidebar if not on a sidebar route', async () => {
    window.history.replaceState('', '', '/project/Solar-Panel-Data-Sorting');

    render(<Project />);

    expect(
      screen.queryByTestId('ProjectSidebar__sidebar'),
    ).not.toBeInTheDocument();
  });

  describe('jobs', () => {
    it('should display job list', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Solar-Panel-Data-Sorting/jobs',
      );

      render(<Project />);

      expect(
        screen.getByTestId('JobListStatic__loadingdots'),
      ).toBeInTheDocument();
      expect(
        await screen.findByTestId('JobList__projectSolar-Panel-Data-Sorting'),
      ).toBeInTheDocument();
    });

    it('should not display logs button', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Solar-Panel-Data-Sorting/jobs',
      );

      render(<Project />);
      expect(screen.queryByText('Read Logs')).not.toBeInTheDocument();
    });

    it('should display logs button', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a',
      );
      render(<Project />);
      const logsLink = await screen.findByRole('link', {name: 'Read Logs'});
      expect(logsLink as HTMLElement).toHaveAttribute(
        'href',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/edges/logs?view=eyJkYXR1bUZpbHRlcnMiOltdfQ%3D%3D',
      );
    });

    it('should display datum logs link with filter applied', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a',
      );
      render(<Project />);
      const logsLink = await waitFor(
        () => screen.findByRole('link', {name: '0 Success'}),
        {timeout: 4000},
      );
      expect(logsLink as HTMLElement).toHaveAttribute(
        'href',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/edges/logs/datum?view=eyJkYXR1bUZpbHRlcnMiOlsiU1VDQ0VTUyJdfQ%3D%3D',
      );
    });

    it('should display loading message for datum info if job is not finshed', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/7798fhje5d4343219bc8e02ff4acd33a',
      );

      render(<Project />);
      expect(await screen.findByText('Processing datums')).toBeInTheDocument();
    });
  });

  describe('pipelines', () => {
    it('should display pipeline details', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Solar-Panel-Data-Sorting/pipelines/montage',
      );

      render(<Project />);

      expect(
        screen.getByTestId('PipelineDetails__pipelineNameSkeleton'),
      ).toBeInTheDocument();

      const pipelineName = await screen.findByTestId('Title__name');

      expect(pipelineName).toHaveTextContent('montage');
    });

    it('should display pipeline logs button', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Solar-Panel-Data-Sorting/pipelines/montage',
      );

      render(<Project />);
      const logsLink = await screen.findByRole('link', {name: 'Inspect Jobs'});
      expect(logsLink as HTMLElement).toHaveAttribute(
        'href',
        '/project/Solar-Panel-Data-Sorting/pipelines/montage/jobs/23b9af7d5d4343219bc8e02ff44cd55a/logs?view=eyJkYXR1bUZpbHRlcnMiOltdfQ%3D%3D',
      );
    });

    it('should display datum logs link with filter applied', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Solar-Panel-Data-Sorting/pipelines/montage',
      );
      render(<Project />);

      const logsLink = await waitFor(
        () => screen.findByRole('link', {name: '2 Success'}),
        {timeout: 4000},
      );
      expect(logsLink as HTMLElement).toHaveAttribute(
        'href',
        '/project/Solar-Panel-Data-Sorting/pipelines/montage/jobs/23b9af7d5d4343219bc8e02ff44cd55a/logs/datum?view=eyJkYXR1bUZpbHRlcnMiOlsiU1VDQ0VTUyJdfQ%3D%3D',
      );
    });

    it('should disable the delete button when there are downstream pipelines', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Egress-Examples/pipelines/edges',
      );

      render(<Project />);
      const deleteButton = await screen.findByTestId(
        'DeletePipelineButton__link',
      );
      expect(deleteButton).toBeDisabled();
    });

    it('should allow pipelines to be deleted', async () => {
      mockServer
        .getState()
        .pipelines['OpenCV-Tutorial'].push(
          mockServer.getState().pipelines['Solar-Panel-Data-Sorting'][0],
        );
      window.history.replaceState(
        '',
        '',
        '/lineage/OpenCV-Tutorial/pipelines/montage',
      );

      render(<Project />);
      expect(mockServer.getState().pipelines['OpenCV-Tutorial']).toHaveLength(
        1,
      );
      const deleteButton = await screen.findByTestId(
        'DeletePipelineButton__link',
      );
      await waitFor(() => expect(deleteButton).toBeEnabled());
      await click(deleteButton);
      const confirmButton = await screen.findByTestId('ModalFooter__confirm');
      await click(confirmButton);

      await waitFor(() =>
        expect(mockServer.getState().pipelines['OpenCV-Tutorial']).toHaveLength(
          0,
        ),
      );
    });

    it('should show job details by default with a globalId filter', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Data-Cleaning-Process/pipelines/likelihoods?view=eyJnbG9iYWxJZEZpbHRlciI6IjIzYjlhZjdkNWQ0MzQzMjE5YmM4ZTAyZmY0YWNkMzNhIn0%3D',
      );

      render(<Project />);

      await screen.findByTestId('InfoPanel__pipeline');
      expect(screen.getByTestId('InfoPanel__pipeline')).toHaveTextContent(
        'likelihoods',
      );
    });
  });

  describe('repos', () => {
    it('should display repo details', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master',
      );

      render(<Project />);

      const repoName = await screen.findByTestId('Title__name');
      const size = screen.getByText('621.86 kB');

      expect(repoName).toHaveTextContent('cron');
      expect(size).toBeInTheDocument();
      await screen.findByText('9d5daa0918ac4c43a476b86e3bb5e88e');
    });

    it('should not show a linked job when there is no job for the commit', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master',
      );

      render(<Project />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('CommitBrowser__loadingdots'),
      );

      await screen.findByTestId('Title__name');

      expect(
        screen.queryByRole('link', {name: 'Linked Job'}),
      ).not.toBeInTheDocument();
    });

    it('should show a linked job for a commit', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Data-Cleaning-Process/repos/models/branch/master/commits',
      );

      render(<Project />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('CommitBrowser__loadingdots'),
      );

      expect(
        await screen.findByRole('link', {name: 'Linked Job'}),
      ).toBeInTheDocument();
    });

    it('should show a linked job for a input repo commit', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Data-Cleaning-Process/repos/training/branch/master/commits',
      );

      render(<Project />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('CommitBrowser__loadingdots'),
      );

      expect(
        await screen.findByRole('link', {name: 'Linked Job'}),
      ).toBeInTheDocument();
    });

    it('should show no commits when the branch has no commits', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Data-Cleaning-Process/repos/training/branch/develop',
      );

      render(<Project />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('CommitBrowser__loadingdots'),
      );

      const emptyMessage = await screen.findByText(
        'There are no commits for this branch',
      );

      expect(emptyMessage).toBeInTheDocument();
    });

    it('should show no branches when the repo has no branches', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Data-Cleaning-Process/repos/test/branch/default',
      );

      render(<Project />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );

      const emptyMessage = await screen.findByText(
        'There are no branches on this repo!',
      );

      expect(emptyMessage).toBeInTheDocument();
    });

    it('should default to the first available branch on repos', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Data-Cleaning-Process/repos/model/branch/default',
      );

      render(<Project />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('CommitBrowser__loadingdots'),
      );

      expect(await screen.findByText('Branch: develop')).toBeInTheDocument();
    });

    it('should show a single commit with diff with a globalId filter', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Data-Cleaning-Process/repos/likelihoods/branch/master?view=eyJnbG9iYWxJZEZpbHRlciI6IjIzYjlhZjdkNWQ0MzQzMjE5YmM4ZTAyZmY0YWNkMzNhIn0%3D',
      );

      render(<Project />);

      await waitFor(() =>
        expect(screen.getByTestId('CommitDetails__id')).toHaveTextContent(
          '23b9af7d5d4343219bc8e02ff4acd33a',
        ),
      );
      expect(
        screen.getByTestId('CommitDetails__fileUpdates'),
      ).toHaveTextContent('1 File updated');
    });

    it('should show empty repo message when repo has no commits', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Data-Cleaning-Process/repos/select/branch/master',
      );

      render(<Project />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('CommitBrowser__loadingdots'),
      );

      const emptyMessage = await screen.findByText(
        'There are no commits for this branch',
      );

      expect(emptyMessage).toBeInTheDocument();
    });

    it('should not display logs button', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master',
      );

      render(<Project />);
      expect(screen.queryByText('Read Logs')).not.toBeInTheDocument();
    });

    it('should disable the delete button when there are associated pipelines', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master',
      );

      render(<Project />);
      const deleteButton = await screen.findByTestId('DeleteRepoButton__link');
      expect(deleteButton).toBeDisabled();
    });

    it('should allow repos to be deleted', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/OpenCV-Tutorial/repos/montage/branch/master',
      );

      render(<Project />);
      expect(mockServer.getState().repos['OpenCV-Tutorial']).toHaveLength(3);
      const deleteButton = await screen.findByTestId('DeleteRepoButton__link');
      await waitFor(() => expect(deleteButton).toBeEnabled());
      await click(deleteButton);
      const confirmButton = await screen.findByTestId('ModalFooter__confirm');
      await click(confirmButton);

      await waitFor(() =>
        expect(mockServer.getState().repos['OpenCV-Tutorial']).toHaveLength(2),
      );
    });

    it('should display a link to pipeline egress', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Egress-Examples/repos/egress_sql/branch/master/info',
      );

      render(
        <Project
          dagLinks={{
            egress_sql_repo: [
              'snowflake://pachyderm@WHMUWUD-CJ80657/PACH_DB/PUBLIC?warehouse=COMPUTE_WH',
            ],
          }}
        />,
      );
      const egress = await screen.findByText(
        'snowflake://pachyderm@WHMUWUD-CJ80657/PACH_DB/PUBLIC?warehouse=COMPUTE_WH',
      );
      await click(egress);
      expect(window.document.execCommand).toHaveBeenCalledWith('copy');
    });

    it('should show a link to view files', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/Data-Cleaning-Process/repos/models/branch/master/commits',
      );

      render(<Project />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('CommitBrowser__loadingdots'),
      );
      expect(await screen.findByText('View Files')).toBeInTheDocument();
      expect(screen.getByText('View Files')).toBeEnabled();
    });

    it('should show a link to view files while filtering for a global id', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Data-Cleaning-Process/repos/likelihoods/branch/default?view=eyJnbG9iYWxJZEZpbHRlciI6IjIzYjlhZjdkNWQ0MzQzMjE5YmM4ZTAyZmY0YWNkMzNhIn0%3D',
      );

      render(<Project />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('CommitDetails__loadingdots'),
      );
      expect(await screen.findByText('View Files')).toBeInTheDocument();
      expect(screen.getByText('View Files')).toBeEnabled();
    });
  });

  it('should filter commits by auto origin', async () => {
    window.history.replaceState(
      '',
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master',
    );

    render(<Project />);

    const hideAutoCommits = await screen.findByLabelText('Auto Commits');
    expect(screen.queryAllByText('View Files')).toHaveLength(6);
    await click(hideAutoCommits);
    await waitFor(() =>
      expect(screen.queryAllByText('View Files')).toHaveLength(2),
    );
  });
});
