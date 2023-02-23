import {
  render,
  waitFor,
  waitForElementToBeRemoved,
  screen,
} from '@testing-library/react';
import React from 'react';

import {
  click,
  mockServer,
  withContextProviders,
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

  describe('pipelines', () => {
    it('should display pipeline details', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Solar-Panel-Data-Sorting/pipelines/montage',
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
        '/lineage/Solar-Panel-Data-Sorting/pipelines/montage',
      );

      render(<Project />);
      const logsLink = await screen.findByRole('link', {name: 'Inspect Jobs'});
      expect(logsLink as HTMLElement).toHaveAttribute(
        'href',
        '/lineage/Solar-Panel-Data-Sorting/pipelines/montage/jobs/23b9af7d5d4343219bc8e02ff44cd55a/logs?view=eyJkYXR1bUZpbHRlcnMiOltdfQ%3D%3D',
      );
    });

    it('should display datum logs link with filter applied', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Solar-Panel-Data-Sorting/pipelines/montage',
      );
      render(<Project />);

      const logsLink = await waitFor(
        () => screen.findByRole('link', {name: '2 Success'}),
        {timeout: 4000},
      );
      expect(logsLink as HTMLElement).toHaveAttribute(
        'href',
        '/lineage/Solar-Panel-Data-Sorting/pipelines/montage/jobs/23b9af7d5d4343219bc8e02ff44cd55a/logs/datum?view=eyJkYXR1bUZpbHRlcnMiOlsiU1VDQ0VTUyJdfQ%3D%3D',
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
        '/lineage/Data-Cleaning-Process/pipelines/likelihoods?view=eyJnbG9iYWxJZEZpbHRlciI6IjIzYjlhZjdkNWQ0MzQzMjE5YmM4ZTAyZmY0YWNkMzNhIn0%3D',
      );

      render(<Project />);

      await screen.findByTestId('InfoPanel__pipeline');
      expect(screen.getByTestId('InfoPanel__pipeline')).toHaveTextContent(
        'likelihoods',
      );
    });

    it('should show a linked project input node', async () => {
      window.history.replaceState(
        {},
        '',
        '/lineage/Multi-Project-Pipeline-A/pipelines/Node_2/job',
      );
      render(<Project />);

      expect(await screen.findByText('Success')).toBeInTheDocument();

      expect(
        await screen.findByText('Node_1 (Project Multi-Project-Pipeline-B)'),
      ).toBeInTheDocument();
    });
  });

  describe('repos', () => {
    it('should display repo details', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master',
      );

      render(<Project />);

      const repoName = await screen.findByTestId('Title__name');
      const size = screen.getByText('543.22 kB');

      expect(repoName).toHaveTextContent('cron');
      expect(size).toBeInTheDocument();
      await screen.findByText('9d5daa0918ac4c43a476b86e3bb5e88e');
      expect(screen.getByText('484.57 kB')).toBeInTheDocument();

      await waitFor(() =>
        expect(
          screen.queryByTestId('CommitList__loadingdots'),
        ).not.toBeInTheDocument(),
      );

      expect(screen.queryAllByTestId('CommitList__commit')).toHaveLength(5);
      expect(screen.getByText('0918a...@master | user')).toBeInTheDocument();
      expect(screen.getByText('0518a...@master | auto')).toBeInTheDocument();
    });

    it('should filter commits by originKind', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master',
      );
      render(<Project />);

      await screen.findByText('9d5daa0918ac4c43a476b86e3bb5e88e');
      await waitFor(() =>
        expect(
          screen.queryByTestId('CommitList__loadingdots'),
        ).not.toBeInTheDocument(),
      );

      expect(screen.queryAllByTestId('CommitList__commit')).toHaveLength(5);
      expect(screen.getByText('0918a...@master | user')).toBeInTheDocument();
      expect(screen.getByText('0518a...@master | auto')).toBeInTheDocument();

      await click(screen.getAllByText('All origins')[0]);
      await click(screen.getByText('User commits only'));

      await waitFor(() =>
        expect(
          screen.queryByTestId('CommitList__loadingdots'),
        ).not.toBeInTheDocument(),
      );

      expect(screen.getByTestId('CommitList__commit')).toBeInTheDocument();
      expect(screen.getByText('0918a...@master | user')).toBeInTheDocument();
      expect(
        screen.queryByText('0518a...@master | auto'),
      ).not.toBeInTheDocument();
    });

    it('should show no branches when the repo has no branches', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Data-Cleaning-Process/repos/test/branch/default',
      );

      render(<Project />);

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );

      const emptyMessage = await screen.findByText(
        `This repo doesn't have any branches`,
      );

      expect(emptyMessage).toBeInTheDocument();
    });

    it('should not display logs button', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master',
      );

      render(<Project />);
      expect(screen.queryByText('Read Logs')).not.toBeInTheDocument();
    });

    it('should disable the delete button when there are associated pipelines', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master',
      );

      render(<Project />);
      const deleteButton = await screen.findByTestId('DeleteRepoButton__link');
      expect(deleteButton).toBeDisabled();
    });

    it('should allow repos to be deleted', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/OpenCV-Tutorial/repos/montage/branch/master',
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
  });
});
