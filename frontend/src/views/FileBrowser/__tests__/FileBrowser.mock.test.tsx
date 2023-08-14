import {
  render,
  screen,
  waitFor,
  within,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import React from 'react';

import {click, withContextProviders} from '@dash-frontend/testHelpers';

import FileBrowserComponent from '../FileBrowser';
const FileBrowser = withContextProviders(() => {
  return <FileBrowserComponent />;
});

window.open = jest.fn();

describe('File Browser', () => {
  // TODO: messed up after removing the test paging size. Will re-write with MSW.
  describe.skip('Download', () => {
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
