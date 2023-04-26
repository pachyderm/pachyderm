import {
  render,
  waitFor,
  waitForElementToBeRemoved,
  screen,
  within,
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
      expect(logsLink).toHaveAttribute(
        'href',
        '/lineage/Solar-Panel-Data-Sorting/pipelines/montage/jobs/23b9af7d5d4343219bc8e02ff44cd55a/logs',
      );
    });

    it('should display datum logs link with filter applied', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Solar-Panel-Data-Sorting/pipelines/montage',
      );
      render(<Project />);

      const logsLink = await screen.findByRole('link', {name: '2 Success'});
      expect(logsLink).toHaveAttribute(
        'href',
        '/lineage/Solar-Panel-Data-Sorting/pipelines/montage/jobs/23b9af7d5d4343219bc8e02ff44cd55a/logs/datum?datumFilters=SUCCESS',
      );
    });

    it('should show a link to file browser for most recent commit', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Data-Cleaning-Process/repos/training/branch/default',
      );

      render(<Project />);

      const fileBrowserLink = await screen.findByRole('link', {
        name: 'Inspect Commits',
      });
      expect(fileBrowserLink).toHaveAttribute(
        'href',
        '/lineage/Data-Cleaning-Process/repos/training/branch/master/commit/23b9af7d5d4343219bc8e02ff4acd33a/?prevPath=%2Flineage%2FData-Cleaning-Process%2Frepos%2Ftraining%2Fbranch%2Fdefault',
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

      const previousCommits = screen.queryAllByTestId('CommitList__commit');
      expect(previousCommits).toHaveLength(5);
      expect(previousCommits[0]).toHaveTextContent(/0918a...@master/);
      expect(previousCommits[4]).toHaveTextContent(/0518a...@master/);
      expect(
        within(previousCommits[0]).getByRole('link', {
          name: 'Inspect Commit',
        }),
      ).toHaveAttribute(
        'href',
        '/lineage/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac4c43a476b86e3bb5e88e9d5daa/?prevPath=%2Flineage%2FSolar-Power-Data-Logger-Team-Collab%2Frepos%2Fcron%2Fbranch%2Fmaster',
      );
      expect(
        within(previousCommits[1]).getByRole('link', {
          name: 'Inspect Commit',
        }),
      ).toHaveAttribute(
        'href',
        '/lineage/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/?prevPath=%2Flineage%2FSolar-Power-Data-Logger-Team-Collab%2Frepos%2Fcron%2Fbranch%2Fmaster',
      );
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

    it('should display a link to repo outputs', async () => {
      window.history.replaceState(
        '',
        '',
        '/lineage/Egress-Examples/repos/edges/branch/default',
      );

      render(
        <Project
          pipelineOutputsMap={{
            'Egress-Examples_edges': [
              {
                id: 'edges_output',
                name: 'edges_output',
              },
              {
                id: 'egress_output',
                name: 'egress_output',
              },
            ],
          }}
        />,
      );

      await waitForElementToBeRemoved(() =>
        screen.queryByTestId('RepoDetails__repoNameSkeleton'),
      );

      expect(screen.getByText('edges_output')).toBeInTheDocument();
      expect(screen.getByText('egress_output')).toBeInTheDocument();
    });
  });
});
