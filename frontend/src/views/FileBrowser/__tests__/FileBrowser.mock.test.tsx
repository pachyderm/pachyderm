import {
  render,
  screen,
  waitFor,
  within,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import React from 'react';

import {
  click,
  withContextProviders,
  mockServer,
} from '@dash-frontend/testHelpers';

import FileBrowserComponent from '../FileBrowser';
const FileBrowser = withContextProviders(() => {
  return <FileBrowserComponent />;
});

window.open = jest.fn();

describe('File Browser', () => {
  describe('Middle Section', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4',
      );
    });

    it('should display commit id and branch, and copy path on click', async () => {
      render(<FileBrowser />);

      expect(
        screen.getByText('Commit: 0918ac9d5daa76b86e3bb5e88e4c43a4'),
      ).toBeInTheDocument();
      expect(screen.getByText('Branch: master')).toBeInTheDocument();

      const copyAction = await screen.findByRole('button', {
        name: 'Copy commit id',
      });
      await click(copyAction);

      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
        'cron@0918ac9d5daa76b86e3bb5e88e4c43a4',
      );
    });

    it('should display file info per table row', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      const files = screen.getAllByTestId('FileTableRow__row');
      expect(files[0]).toHaveTextContent('AT-AT.png');
      expect(files[0]).toHaveTextContent('-');
      expect(files[0]).toHaveTextContent('80.59 kB');
      expect(files[1]).toHaveTextContent('liberty.png');
      expect(files[1]).toHaveTextContent('Added');
      expect(files[1]).toHaveTextContent('58.65 kB');
    });

    it('should navigate to preview on file link', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      await click(screen.getByText('AT-AT.png'));

      expect(window.location.pathname).toBe(
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/AT-AT.png/',
      );
    });

    it('should navigate to dir on file link', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      await click(screen.getByText('cats'));

      expect(window.location.pathname).toBe(
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/cats%2F/',
      );
    });

    it('should navigate up a folder on back button click', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Egress-Examples/repos/images/branch/master/commit/d350c8d08a644ed5b2ee98c035ab6b34/Lorem%2Fipsum%2Fdolor%2F',
      );
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
      expect(screen.getByText('Folder: Lorem/ipsum/dolor')).toBeInTheDocument();

      await click(screen.getByRole('button', {name: 'Back'}));

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      expect(screen.getByText('Folder: Lorem/ipsum')).toBeInTheDocument();
      expect(window.location.pathname).toBe(
        '/project/Egress-Examples/repos/images/branch/master/commit/d350c8d08a644ed5b2ee98c035ab6b34/Lorem%2Fipsum%2F/',
      );

      await click(screen.getByRole('button', {name: 'Back'}));

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      expect(screen.getByText('Folder: Lorem')).toBeInTheDocument();
      expect(window.location.pathname).toBe(
        '/project/Egress-Examples/repos/images/branch/master/commit/d350c8d08a644ed5b2ee98c035ab6b34/Lorem%2F/',
      );
    });

    it('should navigate to file preview on action click', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      await click((await screen.findAllByText('Preview'))[0]);

      expect(window.location.pathname).toBe(
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/AT-AT.png/',
      );
    });

    it('should copy path on action click', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      await click((await screen.findAllByText('Copy Path'))[0]);

      expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
        'cron@master=0918ac9d5daa76b86e3bb5e88e4c43a4:/AT-AT.png',
      );
    });

    it('should delete file on action click', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      expect(
        mockServer.getState().files['Solar-Power-Data-Logger-Team-Collab']['/'],
      ).toHaveLength(20);

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      await click((await screen.findAllByText('Delete'))[0]);

      const deleteConfirm = await screen.findByTestId('ModalFooter__confirm');

      await click(deleteConfirm);

      await waitFor(() =>
        expect(
          mockServer.getState().files['Solar-Power-Data-Logger-Team-Collab'][
            '/'
          ],
        ).toHaveLength(19),
      );
    });

    it.skip('should delete multiple files on action click', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      expect(
        mockServer.getState().files['Solar-Power-Data-Logger-Team-Collab']['/'],
      ).toHaveLength(20);

      const deleteButton = screen.getByRole('button', {
        name: /delete selected items/i,
      });

      expect(deleteButton).toBeDisabled();

      await click(
        screen.getByRole('cell', {
          name: /at-at\.png/i,
        }),
      );
      await click(
        screen.getByRole('cell', {
          name: /cats/i,
        }),
      );
      await click(
        screen.getByRole('cell', {
          name: /json_nested_arrays\.json/i,
        }),
      );

      await waitFor(() => {
        expect(deleteButton).toBeEnabled();
      });

      await click(deleteButton);

      const modal = await screen.findByRole('dialog');
      expect(await within(modal).findByRole('list')).toHaveTextContent(
        '/AT-AT.png/cats//json_nested_arrays.json',
      );
      const deleteConfirm = await within(modal).findByRole('button', {
        name: /delete/i,
      });
      await click(deleteConfirm);

      await waitFor(() =>
        expect(
          mockServer.getState().files['Solar-Power-Data-Logger-Team-Collab'][
            '/'
          ],
        ).toHaveLength(17),
      );
    });

    it('should not allow file deletion for outputRepos', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/processor/branch/master/commit/f4e23cf347c342d98bd9015e4c3ad52a',
      );
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      await click((await screen.findAllByTestId('DropdownButton__button'))[0]);
      expect(screen.queryByText('Delete')).not.toBeInTheDocument();
    });

    it('should allow users to navigate through paged files', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      let files = screen.getAllByTestId('FileTableRow__row');
      expect(files).toHaveLength(15);
      expect(files[0]).toHaveTextContent('AT-AT.png');
      expect(files[14]).toHaveTextContent('txt_spec.txt');

      let pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();
      await click(within(pager).getByTestId('Pager__forward'));

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      files = screen.getAllByTestId('FileTableRow__row');
      expect(files).toHaveLength(5);
      expect(files[0]).toHaveTextContent('xml_plants.xml');
      expect(files[4]).toHaveTextContent('yml_spec_too_large.yml');

      pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__forward')).toBeDisabled();
      await click(within(pager).getByTestId('Pager__backward'));

      files = screen.getAllByTestId('FileTableRow__row');
      expect(files).toHaveLength(15);
    });

    it('should allow users to update page size', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      let files = screen.getAllByTestId('FileTableRow__row');
      expect(files).toHaveLength(15);

      let pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__forward')).toBeEnabled();
      expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();

      await click(within(pager).getByTestId('DropdownButton__button'));
      await click(within(pager).getByText(100));

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      pager = screen.getByTestId('Pager__pager');
      expect(within(pager).getByTestId('Pager__forward')).toBeDisabled();
      expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();

      files = screen.getAllByTestId('FileTableRow__row');
      expect(files).toHaveLength(20);
    });
    it('should display a message if the selected commit is open', async () => {
      window.history.replaceState(
        {},
        '',
        '/lineage/Data-Cleaning-Process/repos/training/branch/test/commit/d4280503cdb44fb984f07952eae8c1ac/',
      );
      render(<FileBrowser />);
      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
      expect(
        screen.getByText('This commit is currently open'),
      ).toBeInTheDocument();
    });
  });

  describe('Right Panel', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4',
      );
    });

    it('should display repo details at top level and in a folder', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
      expect(await screen.findByText('added mako')).toBeInTheDocument();
      expect(await screen.findByText('cron')).toBeInTheDocument();
      expect(await screen.findByText('cron job')).toBeInTheDocument();
      await click(screen.getByText('cats'));
      expect(await screen.findByText('added mako')).toBeInTheDocument();
      expect(await screen.findByText('cron')).toBeInTheDocument();
      expect(await screen.findByText('cron job')).toBeInTheDocument();
    });

    it('should display file history when previewing a file', async () => {
      render(<FileBrowser />);

      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

      await click(screen.getByText('AT-AT.png'));
      expect(await screen.findByText('File Versions')).toBeInTheDocument();
      expect(
        await screen.findByRole('button', {
          name: 'Load older file versions',
        }),
      ).toBeInTheDocument();
    });
  });

  describe('Download', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4',
      );
    });
    afterAll(() => {
      process.env.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_HOST = 'localhost';
      process.env.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_TLS = 'false';
    });

    describe('Port Forward', () => {
      beforeAll(() => {
        process.env.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_HOST = '';
      });

      it('should disable multi file download', async () => {
        render(<FileBrowser />);

        await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

        const downloadButton = screen.getByRole('button', {
          name: /download selected items/i,
        });

        expect(downloadButton).toBeDisabled();

        await click(
          screen.getByRole('cell', {
            name: /at-at\.png/i,
          }),
        );
        await click(
          screen.getByRole('cell', {
            name: /cats/i,
          }),
        );
        await click(
          screen.getByRole('cell', {
            name: /json_nested_arrays\.json/i,
          }),
        );
        expect(downloadButton).toBeDisabled();
      });

      it('should disable single file download for large files', async () => {
        process.env.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_HOST = '';

        render(<FileBrowser />);

        await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

        const pager = screen.getByTestId('Pager__pager');
        expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();
        await click(within(pager).getByTestId('Pager__forward'));

        await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

        await click(
          (
            await screen.findAllByTestId('DropdownButton__button')
          )[3],
        );

        const downloadButton = await screen.findByText(
          'Download (File too large to download)',
        );
        expect(downloadButton.closest('button')).toBeDisabled();
      });
    });

    describe('Proxy', () => {
      beforeEach(() => {
        process.env.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_HOST = 'localhost';
        process.env.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_TLS = '';
      });

      it('should download multiple files on action click', async () => {
        const spy = jest.spyOn(window, 'open');

        render(<FileBrowser />);

        await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

        const downloadButton = screen.getByRole('button', {
          name: /download selected items/i,
        });

        expect(downloadButton).toBeDisabled();

        await click(
          screen.getByRole('cell', {
            name: /at-at\.png/i,
          }),
        );
        await click(
          screen.getByRole('cell', {
            name: /cats/i,
          }),
        );
        await click(
          screen.getByRole('cell', {
            name: /json_nested_arrays\.json/i,
          }),
        );
        expect(downloadButton).toBeEnabled();
        await click(downloadButton);
        await waitFor(() =>
          expect(spy).toHaveBeenLastCalledWith(
            'http://localhost/archive/ASi1L_1gCACFAwASxxgccGkVzh2qIdHkwu39OEmhIn2LktaaoJvq_79nA5Hlg8jGu--4nXKjE1k-SM6348UFshpdDRqkJ6mlVWoVxpJdFFsEv1Mc064CDIAQuRtih5ALqvgu7X2PD_p5g_YyGKF9yIkCAEkFqhViaDgD.zip?authn-token=1',
          ),
        );
      });
      it('should allow download for large files using zip', async () => {
        const spy = jest.spyOn(window, 'open');

        render(<FileBrowser />);

        await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

        const pager = screen.getByTestId('Pager__pager');
        expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();
        await click(within(pager).getByTestId('Pager__forward'));

        await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

        await click(
          (
            await screen.findAllByTestId('DropdownButton__button')
          )[3],
        );

        await click(await screen.findByText('Download Zip'));
        await waitFor(() =>
          expect(spy).toHaveBeenLastCalledWith(
            'http://localhost/archive/ASi1L_0gYd0CABLGFRtwaxMeIippme3f279RC021H2IiEQ5BAlRVTwTZHsDZOpOqk9YbjpTtQQjICVroiFeUSr6gWgZ2ozTMMgqYQIGhuxDlTDVc8xe3Z0wRZ8eH2z80cetURQA.zip?authn-token=1',
          ),
        );
      });

      it('should generate a download link with TLS enabled', async () => {
        process.env.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_TLS = 'true';

        const spy = jest.spyOn(window, 'open');

        render(<FileBrowser />);

        await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

        const pager = screen.getByTestId('Pager__pager');
        expect(within(pager).getByTestId('Pager__backward')).toBeDisabled();
        await click(within(pager).getByTestId('Pager__forward'));

        await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

        await click(
          (
            await screen.findAllByTestId('DropdownButton__button')
          )[3],
        );

        await click(await screen.findByText('Download Zip'));
        await waitFor(() =>
          expect(spy).toHaveBeenLastCalledWith(
            'https://localhost/archive/ASi1L_0gYd0CABLGFRtwaxMeIippme3f279RC021H2IiEQ5BAlRVTwTZHsDZOpOqk9YbjpTtQQjICVroiFeUSr6gWgZ2ozTMMgqYQIGhuxDlTDVc8xe3Z0wRZ8eH2z80cetURQA.zip?authn-token=1',
          ),
        );
      });
    });
  });
});
