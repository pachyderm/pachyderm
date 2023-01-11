import {
  render,
  waitFor,
  waitForElementToBeRemoved,
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
    window.history.replaceState('', '', '/project/1');

    const {queryByTestId} = render(<Project />);

    expect(queryByTestId('ProjectSidebar__sidebar')).not.toBeInTheDocument();
  });

  describe('jobs', () => {
    it('should display job list', async () => {
      window.history.replaceState('', '', '/lineage/1/jobs');

      const {queryByTestId, findByTestId} = render(<Project />);

      expect(queryByTestId('JobListStatic__loadingdots')).toBeInTheDocument();
      expect(await findByTestId('JobList__project1')).toBeInTheDocument();
    });

    it('should not display logs button', async () => {
      window.history.replaceState('', '', '/project/1/jobs');
      const {queryByText} = render(<Project />);
      expect(queryByText('Read Logs')).toBeNull();
    });

    it('should display logs button', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a',
      );
      const {findByRole} = render(<Project />);
      const logsLink = await findByRole('link', {name: 'Read Logs'});
      expect(logsLink as HTMLElement).toHaveAttribute(
        'href',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/edges/logs?view=eyJkYXR1bUZpbHRlcnMiOltdfQ%3D%3D',
      );
    });

    it('should display datum logs link with filter applied', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a',
      );
      const {findByRole} = render(<Project />);
      const logsLink = await waitFor(
        () => findByRole('link', {name: '0 Processed'}),
        {timeout: 4000},
      );
      expect(logsLink as HTMLElement).toHaveAttribute(
        'href',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/edges/logs/datum?view=eyJkYXR1bUZpbHRlcnMiOlsiU1VDQ0VTUyJdfQ%3D%3D',
      );
    });
  });

  describe('pipelines', () => {
    it('should display pipeline details', async () => {
      window.history.replaceState('', '', '/project/1/pipelines/montage');

      const {queryByTestId, findByTestId} = render(<Project />);

      expect(
        queryByTestId('PipelineDetails__pipelineNameSkeleton'),
      ).toBeInTheDocument();

      const pipelineName = await findByTestId('Title__name');

      expect(pipelineName).toHaveTextContent('montage');
    });

    it('should display logs button', async () => {
      window.history.replaceState('', '', '/project/1/pipelines/montage');

      const {findByRole} = render(<Project />);
      const logsLink = await findByRole('link', {name: 'Inspect Jobs'});
      expect(logsLink as HTMLElement).toHaveAttribute(
        'href',
        '/project/1/pipelines/montage/jobs/23b9af7d5d4343219bc8e02ff44cd55a/logs?view=eyJkYXR1bUZpbHRlcnMiOltdfQ%3D%3D',
      );
    });

    it('should display datum logs link with filter applied', async () => {
      window.history.replaceState('', '', '/project/1/pipelines/montage');
      const {findByRole} = render(<Project />);

      const logsLink = await waitFor(
        () => findByRole('link', {name: '0 Processed'}),
        {timeout: 4000},
      );
      expect(logsLink as HTMLElement).toHaveAttribute(
        'href',
        '/project/1/pipelines/montage/jobs/23b9af7d5d4343219bc8e02ff44cd55a/logs/datum?view=eyJkYXR1bUZpbHRlcnMiOlsiU1VDQ0VTUyJdfQ%3D%3D',
      );
    });

    it('should disable the delete button when there are downstream pipelines', async () => {
      window.history.replaceState('', '', '/lineage/5/pipelines/edges');

      const {findByTestId} = render(<Project />);
      const deleteButton = await findByTestId('DeletePipelineButton__link');
      expect(deleteButton).toBeDisabled();
    });

    it('should allow pipelines to be deleted', async () => {
      mockServer
        .getState()
        .pipelines['8'].push(mockServer.getState().pipelines['1'][0]);
      window.history.replaceState('', '', '/lineage/8/pipelines/montage');

      const {findByTestId} = render(<Project />);
      expect(mockServer.getState().pipelines['8']).toHaveLength(1);
      const deleteButton = await findByTestId('DeletePipelineButton__link');
      await waitFor(() => expect(deleteButton).not.toBeDisabled());
      await click(deleteButton);
      const confirmButton = await findByTestId('ModalFooter__confirm');
      await click(confirmButton);

      await waitFor(() =>
        expect(mockServer.getState().pipelines['8']).toHaveLength(0),
      );
    });

    it('should show job details by default with a globalId filter', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/2/pipelines/likelihoods?view=eyJnbG9iYWxJZEZpbHRlciI6IjIzYjlhZjdkNWQ0MzQzMjE5YmM4ZTAyZmY0YWNkMzNhIn0%3D',
      );

      const {getByTestId} = render(<Project />);

      await waitFor(() =>
        expect(getByTestId('InfoPanel__pipeline')).toBeInTheDocument(),
      );
      expect(getByTestId('InfoPanel__pipeline')).toHaveTextContent(
        'likelihoods',
      );
    });
  });

  describe('repos', () => {
    it('should display repo details', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/3/repos/cron/branch/master',
      );

      const {findByTestId, getByText} = render(<Project />);

      const repoName = await findByTestId('Title__name');
      const size = getByText('621.86 kB');

      expect(repoName).toHaveTextContent('cron');
      expect(size).toBeInTheDocument();
      await waitFor(() =>
        expect(
          getByText('9d5daa0918ac4c43a476b86e3bb5e88e'),
        ).toBeInTheDocument(),
      );
    });

    it('should not show a linked job when there is no job for the commit', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/3/repos/cron/branch/master',
      );

      const {findByTestId, queryByRole, getByTestId} = render(<Project />);

      await waitForElementToBeRemoved(() =>
        getByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        getByTestId('CommitBrowser__loadingdots'),
      );

      await findByTestId('Title__name');

      expect(queryByRole('link', {name: 'Linked Job'})).toBeNull();
    });

    it('should show a linked job for a commit', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/2/repos/models/branch/master/commits',
      );

      const {queryByRole, getByTestId} = render(<Project />);

      await waitForElementToBeRemoved(() =>
        getByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        getByTestId('CommitBrowser__loadingdots'),
      );

      await waitFor(() =>
        expect(queryByRole('link', {name: 'Linked Job'})).toBeInTheDocument(),
      );
    });

    it('should show a linked job for a input repo commit', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/2/repos/training/branch/master/commits',
      );

      const {queryByRole, getByTestId} = render(<Project />);

      await waitForElementToBeRemoved(() =>
        getByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        getByTestId('CommitBrowser__loadingdots'),
      );

      await waitFor(() =>
        expect(queryByRole('link', {name: 'Linked Job'})).toBeInTheDocument(),
      );
    });

    it('should show no commits when the branch has no commits', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/2/repos/training/branch/develop',
      );

      const {findByText, getByTestId} = render(<Project />);

      await waitForElementToBeRemoved(() =>
        getByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        getByTestId('CommitBrowser__loadingdots'),
      );

      const emptyMessage = await findByText(
        'There are no commits for this branch',
      );

      expect(emptyMessage).toBeInTheDocument();
    });

    it('should show no branches when the repo has no branches', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/2/repos/test/branch/default',
      );

      const {findByText, getByTestId} = render(<Project />);

      await waitForElementToBeRemoved(() =>
        getByTestId('RepoDetails__repoNameSkeleton'),
      );

      const emptyMessage = await findByText(
        'There are no branches on this repo!',
      );

      expect(emptyMessage).toBeInTheDocument();
    });

    it('should default to the first available branch on repos', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/2/repos/model/branch/default',
      );

      const {findByText, getByTestId} = render(<Project />);

      await waitForElementToBeRemoved(() =>
        getByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        getByTestId('CommitBrowser__loadingdots'),
      );

      expect(await findByText('Branch: develop')).toBeInTheDocument();
    });

    it('should show a single commit with diff with a globalId filter', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/2/repos/likelihoods/branch/master?view=eyJnbG9iYWxJZEZpbHRlciI6IjIzYjlhZjdkNWQ0MzQzMjE5YmM4ZTAyZmY0YWNkMzNhIn0%3D',
      );

      const {getByTestId} = render(<Project />);

      await waitFor(() =>
        expect(getByTestId('CommitDetails__id')).toHaveTextContent(
          '23b9af7d5d4343219bc8e02ff4acd33a',
        ),
      );
      expect(getByTestId('CommitDetails__fileUpdates')).toHaveTextContent(
        '1 File updated',
      );
    });

    it('should show empty repo message when repo has no commits', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/2/repos/select/branch/master',
      );

      const {findByText, getByTestId} = render(<Project />);

      await waitForElementToBeRemoved(() =>
        getByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        getByTestId('CommitBrowser__loadingdots'),
      );

      const emptyMessage = await findByText(
        'There are no commits for this branch',
      );

      expect(emptyMessage).toBeInTheDocument();
    });

    it('should not display logs button', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/3/repos/cron/branch/master',
      );

      const {queryByText} = render(<Project />);
      expect(queryByText('Read Logs')).toBeNull();
    });

    it('should disable the delete button when there are associated pipelines', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/3/repos/cron/branch/master',
      );

      const {findByTestId} = render(<Project />);
      const deleteButton = await findByTestId('DeleteRepoButton__link');
      expect(deleteButton).toBeDisabled();
    });

    it('should allow repos to be deleted', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/8/repos/montage/branch/master',
      );

      const {findByTestId} = render(<Project />);
      expect(mockServer.getState().repos['8']).toHaveLength(3);
      const deleteButton = await findByTestId('DeleteRepoButton__link');
      await waitFor(() => expect(deleteButton).not.toBeDisabled());
      await click(deleteButton);
      const confirmButton = await findByTestId('ModalFooter__confirm');
      await click(confirmButton);

      await waitFor(() =>
        expect(mockServer.getState().repos['8']).toHaveLength(2),
      );
    });

    it('should display a link to pipeline egress', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/5/repos/egress_sql/branch/master/info',
      );

      const {findByText} = render(
        <Project
          dagLinks={{
            egress_sql_repo: [
              'snowflake://pachyderm@WHMUWUD-CJ80657/PACH_DB/PUBLIC?warehouse=COMPUTE_WH',
            ],
          }}
        />,
      );
      const egress = await findByText(
        'snowflake://pachyderm@WHMUWUD-CJ80657/PACH_DB/PUBLIC?warehouse=COMPUTE_WH',
      );
      await click(egress);
      expect(window.document.execCommand).toHaveBeenCalledWith('copy');
    });

    it('should show a link to view files', async () => {
      window.history.replaceState(
        '',
        '',
        '/project/2/repos/models/branch/master/commits',
      );

      const {queryByText, getByTestId} = render(<Project />);

      await waitForElementToBeRemoved(() =>
        getByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        getByTestId('CommitBrowser__loadingdots'),
      );
      expect(queryByText('View Files')).toBeInTheDocument();
      expect(queryByText('View Files')).not.toBeDisabled();
    });

    it('should show a link to view files while filtering for a global id', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/2/repos/likelihoods/branch/default?view=eyJnbG9iYWxJZEZpbHRlciI6IjIzYjlhZjdkNWQ0MzQzMjE5YmM4ZTAyZmY0YWNkMzNhIn0%3D',
      );

      const {queryByText, getByTestId} = render(<Project />);

      await waitForElementToBeRemoved(() =>
        getByTestId('RepoDetails__repoNameSkeleton'),
      );
      await waitForElementToBeRemoved(() =>
        getByTestId('CommitDetails__loadingdots'),
      );
      expect(queryByText('View Files')).toBeInTheDocument();
      expect(queryByText('View Files')).not.toBeDisabled();
    });
  });

  it('should filter commits by auto origin', async () => {
    window.history.replaceState('', '', '/project/3/repos/cron/branch/master');

    const {findByLabelText, queryAllByText} = render(<Project />);

    const hideAutoCommits = await findByLabelText('Auto Commits');
    expect(queryAllByText('View Files')).toHaveLength(6);
    await click(hideAutoCommits);
    await waitFor(() => expect(queryAllByText('View Files')).toHaveLength(2));
  });
});
