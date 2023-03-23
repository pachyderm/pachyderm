import {render, waitFor, act, screen} from '@testing-library/react';
import React from 'react';

import {
  withContextProviders,
  click,
  type,
  mockServer,
} from '@dash-frontend/testHelpers';

import FileBrowserComponent from '../FileBrowser';

const TOTAL_FILES = 17;

describe('File Browser', () => {
  beforeEach(() => {
    window.history.replaceState(
      {},
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4',
    );
  });
  const FileBrowser = withContextProviders(() => {
    return <FileBrowserComponent />;
  });

  describe('File Browser modal', () => {
    it('should display file browser name from url and commit info', async () => {
      render(<FileBrowser />);

      expect(
        await screen.findByText('cron@master=0918ac9d'),
      ).toBeInTheDocument();
    });

    it('should display commit diff info', async () => {
      render(<FileBrowser />);

      expect(await screen.findByText('1 File added')).toBeInTheDocument();
      expect(await screen.findByText('(+58.65 kB)')).toBeInTheDocument();
    });

    it('should filter files', async () => {
      render(<FileBrowser />);

      const searchBar = await screen.findByRole(
        'searchbox',
        {},
        {timeout: 10000},
      );
      expect(await screen.findByText('liberty.png')).toBeInTheDocument();
      expect(await screen.findByText('AT-AT.png')).toBeInTheDocument();
      expect(await screen.findByText('cats')).toBeInTheDocument();
      expect(await screen.findAllByRole('row')).toHaveLength(TOTAL_FILES + 1);

      await type(searchBar, 'lib');

      await waitFor(() => {
        expect(screen.queryByText('AT-AT.png')).not.toBeInTheDocument();
      });
      await waitFor(() => {
        expect(screen.queryByText('cats')).not.toBeInTheDocument();
      });
      expect(screen.getByText('liberty.png')).toBeInTheDocument();
      expect(await screen.findAllByRole('row')).toHaveLength(2);
    });

    it('should show only diff files if toggled', async () => {
      render(<FileBrowser />);

      const diffToggle = await screen.findByRole(
        'switch',
        {name: 'Show diff only'},
        {timeout: 10000},
      );

      expect(await screen.findAllByRole('row')).toHaveLength(TOTAL_FILES + 1);

      await click(diffToggle);

      expect(await screen.findAllByRole('row')).toHaveLength(2);

      await click(diffToggle);

      expect(await screen.findAllByRole('row')).toHaveLength(TOTAL_FILES + 1);
    });

    it('should show an empty state when there are no results', async () => {
      render(<FileBrowser />);

      const searchBar = await screen.findByRole(
        'searchbox',
        {},
        {timeout: 10000},
      );
      await type(searchBar, 'notafile');

      expect(
        await screen.findByText('No Matching Results Found.'),
      ).toBeInTheDocument();
    });

    it('should switch views', async () => {
      render(<FileBrowser />);

      const listViewIcon = await screen.findByLabelText('switch to list view');
      const iconViewIcon = await screen.findByLabelText('switch to icon view');

      await screen.findByTestId('ListViewTable__view');
      expect(
        screen.queryByTestId('FileBrowser__iconView'),
      ).not.toBeInTheDocument();

      await act(async () => {
        await click(iconViewIcon);
      });

      expect(
        screen.queryByTestId('ListViewTable__view'),
      ).not.toBeInTheDocument();
      expect(screen.getByTestId('FileBrowser__iconView')).toBeInTheDocument();

      await click(listViewIcon);

      expect(screen.getByTestId('ListViewTable__view')).toBeInTheDocument();
      expect(
        screen.queryByTestId('FileBrowser__iconView'),
      ).not.toBeInTheDocument();
    });
  });

  describe('File Browser list view', () => {
    it('should display file info', async () => {
      render(<FileBrowser />);

      expect(await screen.findByText('liberty.png')).toBeInTheDocument();
      expect(await screen.findByText('58.65 kB')).toBeInTheDocument();
    });

    it('should sort rows based on different headers', async () => {
      render(<FileBrowser />);

      const nameHeader = await screen.findByLabelText(
        'sort by name in descending order',
      );
      await click(nameHeader);

      let rows = await screen.findAllByRole('row');
      expect(rows[1]).toHaveTextContent(/yml_spec\.yml/);
      expect(rows[2]).toHaveTextContent(/xml_plants\.xml/);
      expect(rows[3]).toHaveTextContent(/txt_spec\.txt/);

      const sizeHeader = await screen.findByLabelText(
        'sort by size in descending order',
      );
      await click(sizeHeader);

      rows = await screen.findAllByRole('row');
      expect(rows[1]).toHaveTextContent(/json_single_field\.json/);
      expect(rows[2]).toHaveTextContent(/csv_commas\.csv/);
      expect(rows[3]).toHaveTextContent(/csv_tabs\.csv/);

      const typeHeader = await screen.findByLabelText(
        'sort by type in descending order',
      );
      await click(typeHeader);

      rows = await screen.findAllByRole('row');
      expect(rows[1]).toHaveTextContent(/cats/);
      expect(rows[2]).toHaveTextContent(/csv_commas\.csv/);
      expect(rows[3]).toHaveTextContent(/csv_tabs\.csv/);
    });

    it('should navigate to dir path on action click', async () => {
      render(<FileBrowser />);

      const seeFilesAction = await screen.findAllByText('See Files');
      await click(seeFilesAction[0]);

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/cats%2F',
        ),
      );

      expect(await screen.findByText('kitten.png')).toBeInTheDocument();
    });

    it('should navigate to file preview on action click', async () => {
      render(<FileBrowser />);

      const previewAction = await screen.findAllByText('Preview');
      await click(previewAction[0]);

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/AT-AT.png',
        ),
      );
    });

    it('should copy path on action click', async () => {
      render(<FileBrowser />);

      const copyAction = await screen.findAllByText('Copy Path');
      await click(copyAction[0]);

      expect(window.document.execCommand).toHaveBeenCalledWith('copy');
    });

    it('should delete a file on delete button click', async () => {
      render(<FileBrowser />);

      const deleteButton = await screen.findAllByTestId(
        'DeleteFileButton__link',
      );
      expect(
        mockServer.getState().files['Solar-Power-Data-Logger-Team-Collab']['/'],
      ).toHaveLength(TOTAL_FILES);

      await click(deleteButton[0]);

      const deleteConfirm = await screen.findByTestId('ModalFooter__confirm');

      await click(deleteConfirm);

      await waitFor(() =>
        expect(
          mockServer.getState().files['Solar-Power-Data-Logger-Team-Collab'][
            '/'
          ],
        ).toHaveLength(TOTAL_FILES - 1),
      );
    });

    it('should not allow file deletion for outputRepos', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/processor/branch/master/commit/f4e23cf347c342d98bd9015e4c3ad52a',
      );
      render(<FileBrowser />);

      expect(
        screen.queryByTestId('DeleteFileButton__link'),
      ).not.toBeInTheDocument();
    });
    it('should not allow preview or download for large file', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/cats%2F',
      );
      render(<FileBrowser />);

      const previewButton = await screen.findByRole('button', {
        name: 'Preview',
      });
      expect(previewButton).toBeDisabled();

      const downloadButton = await screen.findByRole('button', {
        name: 'Download',
      });
      expect(downloadButton).toBeDisabled();
    });
  });

  describe('File Browser icon view', () => {
    it('should display file info', async () => {
      render(<FileBrowser />);

      const iconViewIcon = await screen.findByLabelText('switch to icon view');
      await click(iconViewIcon);

      expect(await screen.findByText('liberty.png')).toBeInTheDocument();
      expect(await screen.findByText('Size: 58.65 kB')).toBeInTheDocument();
    });

    it('should navigate to dir path on action click', async () => {
      render(<FileBrowser />);

      const iconViewIcon = await screen.findByLabelText('switch to icon view');
      await click(iconViewIcon);

      const seeFilesAction = await screen.findAllByText('See Files');
      await click(seeFilesAction[0]);

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/cats%2F',
        ),
      );
    });

    it('should navigate to file preview on action click', async () => {
      render(<FileBrowser />);

      const iconViewIcon = await screen.findByLabelText('switch to icon view');
      await click(iconViewIcon);

      const previewAction = await screen.findAllByText('Preview');
      await click(previewAction[0]);

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/liberty.png',
        ),
      );
    });

    it('should copy path on icon click', async () => {
      render(<FileBrowser />);

      const iconViewIcon = await screen.findByLabelText('switch to icon view');

      await click(iconViewIcon);

      const copyAction = await screen.findAllByLabelText('Copy');
      await click(copyAction[0]);

      expect(window.document.execCommand).toHaveBeenCalledWith('copy');
    });

    it('should delete a file on delete icon click', async () => {
      render(<FileBrowser />);

      const deleteButton = await screen.findAllByTestId(
        'DeleteFileButton__link',
      );
      expect(
        mockServer.getState().files['Solar-Power-Data-Logger-Team-Collab']['/'],
      ).toHaveLength(TOTAL_FILES);

      await click(deleteButton[0]);

      const deleteConfirm = await screen.findByTestId('ModalFooter__confirm');

      await click(deleteConfirm);

      await waitFor(() =>
        expect(
          mockServer.getState().files['Solar-Power-Data-Logger-Team-Collab'][
            '/'
          ],
        ).toHaveLength(TOTAL_FILES - 1),
      );
    });

    it('should not allow file deletion for outputRepos', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/processor/branch/master/commit/f4e23cf347c342d98bd9015e4c3ad52a',
      );
      render(<FileBrowser />);

      expect(
        screen.queryByTestId('DeleteFileButton__link'),
      ).not.toBeInTheDocument();
    });

    it('should not allow preview or download for large file', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/cats%2F',
      );
      render(<FileBrowser />);

      const iconViewIcon = await screen.findByLabelText('switch to icon view');
      await click(iconViewIcon);

      const previewButton = await screen.findByRole('button', {
        name: 'Preview',
      });
      expect(previewButton).toBeDisabled();

      const downloadButton = screen.getByRole('button', {
        name: 'Download /cats/test.png',
      });
      expect(downloadButton).toBeDisabled();
    });
  });

  describe('File Browser file preview', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/liberty.png',
      );
    });

    it('should show file preview', async () => {
      render(<FileBrowser />);

      expect(
        await screen.findByText(/Uploaded: Jan 8, 2021/),
      ).toBeInTheDocument();
      expect(await screen.findByText(/(\+58.65 kB)/)).toBeInTheDocument();
      expect(await screen.findByRole('img')).toHaveAttribute(
        'src',
        //download/images/master/d350c8d08a644ed5b2ee98c035ab6b33/liberty.png/,
      );
    });

    it('should go to path based on breadcrumb', async () => {
      render(<FileBrowser />);

      const topButton = await screen.findByTestId('Breadcrumb__home');
      await click(topButton);

      expect(
        await screen.findByLabelText('switch to list view'),
      ).toBeInTheDocument();
    });

    it('should copy path on icon click', async () => {
      render(<FileBrowser />);

      const copyAction = await screen.findByText('Copy Path');
      await click(copyAction);

      expect(window.document.execCommand).toHaveBeenCalledWith('copy');
    });

    it('should delete a file on delete button click', async () => {
      render(<FileBrowser />);

      const deleteButton = await screen.findByTestId('DeleteFileButton__link');
      expect(
        mockServer.getState().files['Solar-Power-Data-Logger-Team-Collab']['/'],
      ).toHaveLength(TOTAL_FILES);

      await click(deleteButton);

      const deleteConfirm = await screen.findByTestId('ModalFooter__confirm');

      await click(deleteConfirm);

      await waitFor(() =>
        expect(
          mockServer.getState().files['Solar-Power-Data-Logger-Team-Collab'][
            '/'
          ],
        ).toHaveLength(TOTAL_FILES - 1),
      );
    });

    it('should not allow file deletion for outputRepos', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/processor/branch/master/commit/f4e23cf347c342d98bd9015e4c3ad52a',
      );
      render(<FileBrowser />);

      expect(
        screen.queryByTestId('DeleteFileButton__link'),
      ).not.toBeInTheDocument();
    });
  });
});
