import {render, waitFor, act} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import {withContextProviders, click} from '@dash-frontend/testHelpers';

import FileBrowserComponent from '../FileBrowser';

describe('File Browser', () => {
  beforeEach(() => {
    window.history.replaceState(
      {},
      '',
      '/project/3/repo/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4',
    );
  });
  const FileBrowser = withContextProviders(() => {
    return <FileBrowserComponent />;
  });
  describe('File Browser modal', () => {
    it('should display file browser name from url', async () => {
      const {findByText} = render(<FileBrowser />);

      expect(await findByText('cron@master=0918ac9d')).toBeInTheDocument();
    });

    it('should filter files', async () => {
      const {queryByText, findByText, findByRole, findAllByRole} = render(
        <FileBrowser />,
      );

      const searchBar = await findByRole('searchbox');
      expect(await findByText('liberty.png')).toBeInTheDocument();
      expect(await findByText('kitten.png')).toBeInTheDocument();
      expect(await findAllByRole('row')).toHaveLength(5);

      userEvent.type(searchBar, 'kit');

      await waitFor(() =>
        expect(queryByText('liberty.png')).not.toBeInTheDocument(),
      );
      expect(queryByText('kitten.png')).toBeInTheDocument();
      expect(await findAllByRole('row')).toHaveLength(2);
    });

    it('should show an empty state when there are no results', async () => {
      const {queryByText, findByRole} = render(<FileBrowser />);

      const searchBar = await findByRole('searchbox');
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
      expect(await findByText('57.27 KB')).toBeInTheDocument();
      expect(await findByText('January 8, 2021')).toBeInTheDocument();
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
      expect(rows[1].textContent).toContain('liberty.png');
      expect(rows[2].textContent).toContain('kitten.png');
      expect(rows[3].textContent).toContain('cats');
      expect(rows[4].textContent).toContain('AT-AT.png');

      const sizeHeader = await findByLabelText(
        'sort by size in descending order',
      );
      await act(async () => {
        userEvent.click(sizeHeader);
      });

      rows = await findAllByRole('row');
      expect(rows[1].textContent).toContain('liberty.png');
      expect(rows[2].textContent).toContain('AT-AT.png');
      expect(rows[3].textContent).toContain('cats');
      expect(rows[4].textContent).toContain('kitten.png');

      const typeHeader = await findByLabelText(
        'sort by type in descending order',
      );
      await act(async () => {
        userEvent.click(typeHeader);
      });

      rows = await findAllByRole('row');
      expect(rows[1].textContent).toContain('cats');
      expect(rows[2].textContent).toContain('AT-AT.png');
      expect(rows[3].textContent).toContain('liberty.png');
      expect(rows[4].textContent).toContain('kitten.png');

      const dateHeader = await findByLabelText(
        'sort by date in descending order',
      );
      await act(async () => {
        userEvent.click(dateHeader);
      });

      rows = await findAllByRole('row');
      expect(rows[1].textContent).toContain('liberty.png');
      expect(rows[2].textContent).toContain('kitten.png');
      expect(rows[3].textContent).toContain('AT-AT.png');
      expect(rows[4].textContent).toContain('cats');
    });

    it('should navigate to dir path on action click', async () => {
      const {findAllByText} = render(<FileBrowser />);

      const seeFilesAction = await findAllByText('See Files');
      click(seeFilesAction[0]);

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/3/repo/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/cats%2F',
        ),
      );
    });

    it('should navigate to file preview on action click', async () => {
      const {findAllByText} = render(<FileBrowser />);

      const previewAction = await findAllByText('Preview');
      click(previewAction[0]);

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/3/repo/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/AT-AT.png',
        ),
      );
    });

    it('should copy path on action click', async () => {
      const {findAllByText} = render(<FileBrowser />);

      const copyAction = await findAllByText('Copy Path');
      await act(async () => click(copyAction[0]));

      expect(window.document.execCommand).toHaveBeenCalledWith('copy');
    });
  });

  describe('File Browser icon view', () => {
    it('should display file info', async () => {
      const {findByText, findByLabelText} = render(<FileBrowser />);

      const iconViewIcon = await findByLabelText('switch to icon view');
      click(iconViewIcon);

      expect(await findByText('liberty.png')).toBeInTheDocument();
      expect(await findByText('Size: 57.27 KB')).toBeInTheDocument();
      expect(await findByText('Uploaded: January 8, 2021')).toBeInTheDocument();
    });

    it('should navigate to dir path on action click', async () => {
      const {findAllByText, findByLabelText} = render(<FileBrowser />);

      const iconViewIcon = await findByLabelText('switch to icon view');
      click(iconViewIcon);

      const seeFilesAction = await findAllByText('See Files');
      click(seeFilesAction[0]);

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/3/repo/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/cats%2F',
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
          '/project/3/repo/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/AT-AT.png',
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
  });

  describe('File Browser file preview', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/project/3/repo/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/liberty.png',
      );
    });

    it('should show file preview', async () => {
      const {findByText, findByRole} = render(<FileBrowser />);

      expect(
        await findByText('Uploaded: January 8, 2021 (57.27 KB)'),
      ).toBeInTheDocument();
      expect(await findByRole('img')).toHaveAttribute(
        'src',
        //download/images/master/d350c8d08a644ed5b2ee98c035ab6b33/liberty.png/,
      );
    });

    it('should go to path based on breadcrumb', async () => {
      const {queryByTestId, findByText} = render(<FileBrowser />);

      const topButton = await findByText('top');
      click(topButton);

      await waitFor(() =>
        expect(queryByTestId('ListViewTable__view')).toBeInTheDocument(),
      );
    });

    it('should copy path on icon click', async () => {
      const {findByText} = render(<FileBrowser />);

      const copyAction = await findByText('Copy Path');
      await act(async () => click(copyAction));

      expect(window.document.execCommand).toHaveBeenCalledWith('copy');
    });
  });
});
