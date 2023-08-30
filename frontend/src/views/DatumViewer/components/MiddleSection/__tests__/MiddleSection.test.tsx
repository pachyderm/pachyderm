import {render, waitFor, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import useDownloadText from '@dash-frontend/hooks/useDownloadText';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {
  mockEmptyGetAuthorize,
  mockGetJob5CDatum05,
  mockGetJob5CDatumCH,
  mockGetMontageJob_5C,
  mockGetMontagePipeline,
  mockGetServicePipeline,
  mockGetSpoutPipeline,
  mockGetVersionInfo,
} from '@dash-frontend/mocks';
import {mockEmptyGetLogs, mockGetLogs} from '@dash-frontend/mocks/logs';
import {withContextProviders, click} from '@dash-frontend/testHelpers';

import {default as MiddleSectionComponent} from '../MiddleSection';

const MiddleSection = withContextProviders(() => {
  return <MiddleSectionComponent />;
});

jest.mock(
  'react-virtualized-auto-sizer',
  () =>
    ({
      children,
    }: {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      children: (size: {height: number; width: number}) => any;
    }) =>
      children({height: 100, width: 50}),
);

jest.mock('@dash-frontend/hooks/useDownloadText');
const mockUseDownloadText = useDownloadText as jest.MockedFunction<
  typeof useDownloadText
>;
const download = jest.fn();
mockUseDownloadText.mockReturnValue({
  download,
  downloaded: false,
  reset: jest.fn(),
});

const basePath =
  '/project/default/jobs/5c1aa9bc87dd411ba5a1be0c80a3ebc2/pipeline/montage/logs';

describe('Datum Viewer Middle Section', () => {
  const server = setupServer();

  beforeAll(() => server.listen());

  beforeEach(() => {
    window.history.replaceState({}, '', basePath);
    window.localStorage.removeItem('pachyderm-console-default');
    server.resetHandlers();
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetVersionInfo());
    server.use(mockGetMontagePipeline());
    server.use(mockGetMontageJob_5C());
    server.use(mockGetJob5CDatum05());
    server.use(mockGetLogs());
  });

  afterAll(() => server.close());

  describe('Logs Viewer', () => {
    it('should display empty state when there are no logs', async () => {
      server.use(mockEmptyGetLogs());
      render(<MiddleSection />);
      await screen.findByText('No logs found for this time range.');
    });

    it('should display logs', async () => {
      render(<MiddleSection />);
      const rows = await screen.findAllByTestId('LogRow__base');
      expect(rows).toHaveLength(7);
      expect(rows[0]).toHaveTextContent(
        `${getStandardDate(1690919093)} started process datum set task`,
      );
      expect(rows[6]).toHaveTextContent(
        `${getStandardDate(1690919095)} finished process datum set task`,
      );
    });

    it('should highlight user logs', async () => {
      render(<MiddleSection />);
      await screen.findAllByTestId('LogRow__base');
      expect(screen.queryByTestId('LogRow__user_log')).not.toBeInTheDocument();

      await click(await screen.findByTestId('DropdownButton__button'));
      await click(screen.getByText('Highlight User Logs'));
      expect(await screen.findAllByTestId('LogRow__user_log')).toHaveLength(3);
    });

    it('should display raw logs', async () => {
      render(<MiddleSection />);

      await click(await screen.findByTestId('DropdownButton__button'));
      await click(screen.getByText('Raw Logs'));

      const rows = await screen.findAllByTestId('RawLogRow__base');
      expect(rows).toHaveLength(7);
      expect(rows[0]).toHaveTextContent('started process datum set task');
      expect(rows[6]).toHaveTextContent('finished process datum set task');
    });

    it('should highlight raw user logs', async () => {
      render(<MiddleSection />);
      await screen.findAllByTestId('LogRow__base');

      await click(await screen.findByTestId('DropdownButton__button'));
      await click(screen.getByText('Raw Logs'));
      expect(
        screen.queryByTestId('RawLogRow__user_log'),
      ).not.toBeInTheDocument();
      await click(screen.getByText('Highlight User Logs'));
      expect(await screen.findAllByTestId('RawLogRow__user_log')).toHaveLength(
        3,
      );
    });

    it('should display message for a skipped datum', async () => {
      server.use(mockGetJob5CDatumCH());
      window.history.replaceState(
        {},
        '',
        `${basePath}/datum/ch3db37fa4594a00ebf1dc972f81b58de642cd0cfca811e1b5bd6a2bb292a8e0`,
      );
      render(<MiddleSection />);

      await waitFor(() => {
        expect(
          screen.getByRole('heading', {
            name: /skipped datum\./i,
          }),
        ).toBeInTheDocument();
      });

      expect(
        await screen.findByText(
          'This datum has been successfully processed in a previous job.',
        ),
      ).toBeInTheDocument();
      expect(
        screen.queryByRole('button', {
          name: /refresh/i,
        }),
      ).not.toBeInTheDocument();
    });

    it('should store logs view preference on refresh', async () => {
      render(<MiddleSection />);

      expect(await screen.findAllByTestId('LogRow__base')).toHaveLength(7);
      expect(screen.queryByTestId('RawLogRow__base')).not.toBeInTheDocument();
      await click(await screen.findByTestId('DropdownButton__button'));
      await click(screen.getByText('Raw Logs'));
      expect(screen.queryByTestId('LogRow__base')).not.toBeInTheDocument();
      expect(await screen.findAllByTestId('RawLogRow__base')).toHaveLength(7);

      window.history.replaceState({}, '', basePath);
      expect(screen.queryByTestId('LogRow__base')).not.toBeInTheDocument();
      expect(await screen.findAllByTestId('RawLogRow__base')).toHaveLength(7);
    });

    it('should show the correct breadcrumbs and title for job logs', async () => {
      render(<MiddleSection />);
      expect(
        await screen.findByTestId('DatumHeaderBreadCrumbs__path'),
      ).toHaveTextContent('Pipeline.../Job: 5c1aa9bc87dd411ba5a1be0c80a3ebc2');

      expect(
        await screen.findByTestId('MiddleSection__title'),
      ).toHaveTextContent('Job Logs for5c1aa9bc87dd411ba5a1be0c80a3ebc2');
    });

    it('should show the correct breadcrumbs and title for datum logs', async () => {
      window.history.replaceState(
        {},
        '',
        `${basePath}/datum/05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7`,
      );
      render(<MiddleSection />);
      expect(
        await screen.findByTestId('DatumHeaderBreadCrumbs__path'),
      ).toHaveTextContent(
        'Pipeline.../Job.../Datum: 05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7',
      );

      expect(
        await screen.findByTestId('MiddleSection__title'),
      ).toHaveTextContent(
        'Datum Logs for05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7',
      );

      const link = screen.getByRole('link', {
        name: 'Job...',
      });
      expect(link).toHaveAttribute('href', basePath);
      await click(link);

      expect(
        await screen.findByTestId('DatumHeaderBreadCrumbs__path'),
      ).toHaveTextContent('Pipeline.../Job: 5c1aa9bc87dd411ba5a1be0c80a3ebc2');
    });

    it('should display correct Spout Pipeline header', async () => {
      server.use(mockGetSpoutPipeline());
      render(<MiddleSection />);
      expect(
        await screen.findByTestId('MiddleSection__title'),
      ).toHaveTextContent('Pipeline logs formontage');
    });

    it('should display correct Service Pipeline header', async () => {
      server.use(mockGetServicePipeline());
      render(<MiddleSection />);
      expect(
        await screen.findByTestId('MiddleSection__title'),
      ).toHaveTextContent('Pipeline logs formontage');
    });

    it('export options should download and copy selected logs', async () => {
      render(<MiddleSection />);

      const downloadButton = await screen.findByRole('button', {
        name: 'Download selected logs',
      });

      const copyButton = await screen.findByRole('button', {
        name: 'Copy selected logs',
      });

      const rows = await screen.findAllByTestId('LogRow__checkbox');

      expect(copyButton).toBeDisabled();
      expect(downloadButton).toBeDisabled();
      await click(rows[3]);

      expect(copyButton).toBeEnabled();
      expect(downloadButton).toBeEnabled();

      const selectedLogs = `${getStandardDate(
        1690919094,
      )} montage: no decode delegate for this image format \`' @ error/constitute.c/ReadImage/740.`;

      expect(navigator.clipboard.writeText).toHaveBeenCalledTimes(0);
      await click(copyButton);
      expect(navigator.clipboard.writeText).toHaveBeenCalledTimes(1);
      expect(navigator.clipboard.writeText).toHaveBeenLastCalledWith(
        selectedLogs,
      );

      expect(download).toHaveBeenCalledTimes(0);
      await click(downloadButton);
      expect(download).toHaveBeenCalledTimes(1);
      expect(mockUseDownloadText).toHaveBeenLastCalledWith(
        selectedLogs,
        'montage_logs',
      );
    });

    it('should show logs pager', async () => {
      render(<MiddleSection />);
      await screen.findAllByTestId('LogRow__base');

      const forwards = await screen.findByTestId('Pager__forward');
      const backwards = await screen.findByTestId('Pager__backward');

      const refresh = await screen.findByRole('button', {
        name: /refresh/i,
      });

      expect(refresh).toBeEnabled();
      expect(backwards).toBeDisabled();
      expect(forwards).toBeDisabled();
    });
  });
});
