import {render, waitFor, act} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import {
  withContextProviders,
  click,
  mockServer,
} from '@dash-frontend/testHelpers';

import FileBrowserComponent from '../FileBrowser';

const TOTAL_FILES = 17;

describe('File Browser', () => {
  beforeEach(() => {
    window.history.replaceState(
      {},
      '',
      '/project/3/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4',
    );
  });
  const FileBrowser = withContextProviders(() => {
    return <FileBrowserComponent />;
  });

  describe('File Browser modal', () => {
    it('should display file browser name from url and commit info', async () => {
      const {findByText} = render(<FileBrowser />);

      expect(await findByText('cron@master=0918ac9d')).toBeInTheDocument();
    });

    it('should display commit diff info', async () => {
      const {findByText} = render(<FileBrowser />);

      expect(await findByText('1 File added')).toBeInTheDocument();
      expect(await findByText('(+58.65 kB)')).toBeInTheDocument();
    });

    it('should filter files', async () => {
      const {queryByText, findByText, findByRole, findAllByRole} = render(
        <FileBrowser />,
      );

      const searchBar = await findByRole('searchbox', {}, {timeout: 10000});
      expect(await findByText('liberty.png')).toBeInTheDocument();
      expect(await findByText('AT-AT.png')).toBeInTheDocument();
      expect(await findByText('cats')).toBeInTheDocument();
      expect(await findAllByRole('row')).toHaveLength(TOTAL_FILES + 1);

      userEvent.type(searchBar, 'lib');

      await waitFor(() => {
        expect(queryByText('AT-AT.png')).not.toBeInTheDocument();
        expect(queryByText('cats')).not.toBeInTheDocument();
      });
      expect(queryByText('liberty.png')).toBeInTheDocument();
      expect(await findAllByRole('row')).toHaveLength(2);
    });

    it('should show only diff files if toggled', async () => {
      const {findByRole, findAllByRole} = render(<FileBrowser />);

      const diffToggle = await findByRole(
        'switch',
        {name: 'Show diff only'},
        {timeout: 10000},
      );

      expect(await findAllByRole('row')).toHaveLength(TOTAL_FILES + 1);

      click(diffToggle);

      expect(await findAllByRole('row')).toHaveLength(2);

      click(diffToggle);

      expect(await findAllByRole('row')).toHaveLength(TOTAL_FILES + 1);
    });

    it('should show an empty state when there are no results', async () => {
      const {queryByText, findByRole} = render(<FileBrowser />);

      const searchBar = await findByRole('searchbox', {}, {timeout: 10000});
      userEvent.type(searchBar, 'notafile');

      await waitFor(() =>
        expect(queryByText('No Matching Results Found.')).toBeInTheDocument(),
      );
    });

    it('should switch views', async () => {
      const {queryByTestId, findByLabelText} = render(<FileBrowser />);

      const listViewIcon = await findByLabelText('switch to list view');
      const iconViewIcon = await findByLabelText('switch to icon view');

      await waitFor(() =>
        expect(queryByTestId('ListViewTable__view')).toBeInTheDocument(),
      );
      expect(queryByTestId('FileBrowser__iconView')).not.toBeInTheDocument();

      await act(async () => {
        click(iconViewIcon);
      });

      expect(queryByTestId('ListViewTable__view')).not.toBeInTheDocument();
      expect(queryByTestId('FileBrowser__iconView')).toBeInTheDocument();

      await act(async () => {
        click(listViewIcon);
      });

      expect(queryByTestId('ListViewTable__view')).toBeInTheDocument();
      expect(queryByTestId('FileBrowser__iconView')).not.toBeInTheDocument();
    });
  });

  describe('File Browser list view', () => {
    it('should display file info', async () => {
      const {findByText} = render(<FileBrowser />);

      expect(await findByText('liberty.png')).toBeInTheDocument();
      expect(await findByText('58.65 kB')).toBeInTheDocument();
    });

    it('should sort rows based on different headers', async () => {
      const {findAllByRole, findByLabelText} = render(<FileBrowser />);

      const nameHeader = await findByLabelText(
        'sort by name in descending order',
      );
      await act(async () => {
        userEvent.click(nameHeader);
      });

      let rows = await findAllByRole('row');
      expect(rows[1].textContent).toContain('yml_spec.yml');
      expect(rows[2].textContent).toContain('xml_plants.xml');
      expect(rows[3].textContent).toContain('txt_spec.txt');

      const sizeHeader = await findByLabelText(
        'sort by size in descending order',
      );
      await act(async () => {
        userEvent.click(sizeHeader);
      });

      rows = await findAllByRole('row');
      expect(rows[1].textContent).toContain('json_single_field.json');
      expect(rows[2].textContent).toContain('csv_commas.csv');
      expect(rows[3].textContent).toContain('csv_tabs.csv');

      const typeHeader = await findByLabelText(
        'sort by type in descending order',
      );
      await act(async () => {
        userEvent.click(typeHeader);
      });

      rows = await findAllByRole('row');
      expect(rows[1].textContent).toContain('cats');
      expect(rows[2].textContent).toContain('csv_commas.csv');
      expect(rows[3].textContent).toContain('csv_tabs.csv');
    });

    it('should navigate to dir path on action click', async () => {
      const {findAllByText, findByText} = render(<FileBrowser />);

      const seeFilesAction = await findAllByText('See Files');
      click(seeFilesAction[0]);

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/3/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/cats%2F',
        ),
      );

      expect(await findByText('kitten.png')).toBeInTheDocument();
    });

    it('should navigate to file preview on action click', async () => {
      const {findAllByText} = render(<FileBrowser />);

      const previewAction = await findAllByText('Preview');
      click(previewAction[0]);

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/3/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/AT-AT.png',
        ),
      );
    });

    it('should copy path on action click', async () => {
      const {findAllByText} = render(<FileBrowser />);

      const copyAction = await findAllByText('Copy Path');
      await act(async () => click(copyAction[0]));

      expect(window.document.execCommand).toHaveBeenCalledWith('copy');
    });

    it('should delete a file on delete button click', async () => {
      const {findByTestId, findAllByTestId} = render(<FileBrowser />);

      const deleteButton = await findAllByTestId('DeleteFileButton__link');
      expect(mockServer.getState().files['3']['/']).toHaveLength(TOTAL_FILES);

      click(deleteButton[0]);

      const deleteConfirm = await findByTestId('ModalFooter__confirm');

      click(deleteConfirm);

      await waitFor(() =>
        expect(mockServer.getState().files['3']['/']).toHaveLength(
          TOTAL_FILES - 1,
        ),
      );
    });

    it('should not allow file deletion for outputRepos', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/3/repos/processor/branch/master/commit/f4e23cf347c342d98bd9015e4c3ad52a',
      );
      const {queryByTestId} = render(<FileBrowser />);

      expect(queryByTestId('DeleteFileButton__link')).not.toBeInTheDocument();
    });
  });

  describe('File Browser icon view', () => {
    it('should display file info', async () => {
      const {findByText, findByLabelText} = render(<FileBrowser />);

      const iconViewIcon = await findByLabelText('switch to icon view');
      click(iconViewIcon);

      expect(await findByText('liberty.png')).toBeInTheDocument();
      expect(await findByText('Size: 58.65 kB')).toBeInTheDocument();
    });

    it('should navigate to dir path on action click', async () => {
      const {findAllByText, findByLabelText} = render(<FileBrowser />);

      const iconViewIcon = await findByLabelText('switch to icon view');
      click(iconViewIcon);

      const seeFilesAction = await findAllByText('See Files');
      click(seeFilesAction[0]);

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/3/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/cats%2F',
        ),
      );
    });

    it('should navigate to file preview on action click', async () => {
      const {findAllByText, findByLabelText} = render(<FileBrowser />);

      const iconViewIcon = await findByLabelText('switch to icon view');
      click(iconViewIcon);

      const previewAction = await findAllByText('Preview');
      click(previewAction[0]);

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/3/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/liberty.png',
        ),
      );
    });

    it('should copy path on icon click', async () => {
      const {findAllByLabelText, findByLabelText} = render(<FileBrowser />);

      const iconViewIcon = await findByLabelText('switch to icon view');

      click(iconViewIcon);

      const copyAction = await findAllByLabelText('Copy');
      await act(async () => userEvent.click(copyAction[0]));

      expect(window.document.execCommand).toHaveBeenCalledWith('copy');
    });

    it('should delete a file on delete icon click', async () => {
      const {findByTestId, findAllByTestId} = render(<FileBrowser />);

      const deleteButton = await findAllByTestId('DeleteFileButton__link');
      expect(mockServer.getState().files['3']['/']).toHaveLength(TOTAL_FILES);

      click(deleteButton[0]);

      const deleteConfirm = await findByTestId('ModalFooter__confirm');

      click(deleteConfirm);

      await waitFor(() =>
        expect(mockServer.getState().files['3']['/']).toHaveLength(
          TOTAL_FILES - 1,
        ),
      );
    });

    it('should not allow file deletion for outputRepos', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/3/repos/processor/branch/master/commit/f4e23cf347c342d98bd9015e4c3ad52a',
      );
      const {queryByTestId} = render(<FileBrowser />);

      expect(queryByTestId('DeleteFileButton__link')).not.toBeInTheDocument();
    });
  });

  describe('File Browser file preview', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/project/3/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/liberty.png',
      );
    });

    it('should show file preview', async () => {
      const {findByText, findByRole} = render(<FileBrowser />);

      expect(
        await findByText('Uploaded: January 8, 2021 (58.65 kB)'),
      ).toBeInTheDocument();
      expect(await findByRole('img')).toHaveAttribute(
        'src',
        //download/images/master/d350c8d08a644ed5b2ee98c035ab6b33/liberty.png/,
      );
    });

    it('should go to path based on breadcrumb', async () => {
      const {queryByLabelText, findByTestId} = render(<FileBrowser />);

      const topButton = await findByTestId('Breadcrumb__home');
      click(topButton);

      await waitFor(() =>
        expect(queryByLabelText('switch to list view')).toBeInTheDocument(),
      );
    });

    it('should copy path on icon click', async () => {
      const {findByText} = render(<FileBrowser />);

      const copyAction = await findByText('Copy Path');
      await act(async () => click(copyAction));

      expect(window.document.execCommand).toHaveBeenCalledWith('copy');
    });

    it('should delete a file on delete button click', async () => {
      const {findByTestId} = render(<FileBrowser />);

      const deleteButton = await findByTestId('DeleteFileButton__link');
      expect(mockServer.getState().files['3']['/']).toHaveLength(TOTAL_FILES);

      click(deleteButton);

      const deleteConfirm = await findByTestId('ModalFooter__confirm');

      click(deleteConfirm);

      await waitFor(() =>
        expect(mockServer.getState().files['3']['/']).toHaveLength(
          TOTAL_FILES - 1,
        ),
      );
    });

    it('should not allow file deletion for outputRepos', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/3/repos/processor/branch/master/commit/f4e23cf347c342d98bd9015e4c3ad52a',
      );
      const {queryByTestId} = render(<FileBrowser />);

      expect(queryByTestId('DeleteFileButton__link')).not.toBeInTheDocument();
    });
  });
});
