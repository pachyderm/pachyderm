import {
  render,
  waitFor,
  within,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import {useClipboardCopy} from '@dash-frontend/../components/src';
import useDownloadText from '@dash-frontend/hooks/useDownloadText';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {withContextProviders, click, type} from '@dash-frontend/testHelpers';

import {default as MiddleSectionComponent} from '../components/MiddleSection';
import {
  PipelineDatumViewer as PipelineDatumViewerComponent,
  JobDatumViewer as JobDatumViewerComponent,
} from '../DatumViewer';
const PipelineDatumViewer = withContextProviders(() => {
  return <PipelineDatumViewerComponent />;
});
const JobDatumViewer = withContextProviders(() => {
  return <JobDatumViewerComponent />;
});

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

jest.mock('@pachyderm/components', () => {
  const original = jest.requireActual('@pachyderm/components');
  return {
    ...original,
    useClipboardCopy: jest.fn(),
  };
});
const mockUseClipboardCopy = useClipboardCopy as jest.MockedFunction<
  typeof useClipboardCopy
>;
const copy = jest.fn();
mockUseClipboardCopy.mockReturnValue({
  copy,
  supported: true,
  copied: false,
  reset: jest.fn(),
});

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

describe('Datum Viewer', () => {
  describe('on close', () => {
    it('should route user back to pipeline view on close', async () => {
      window.history.replaceState(
        {},
        '',
        '/lineage/Solar-Panel-Data-Sorting/pipelines/montage/jobs/23b9af7d5d4343219bc8e02ff44cd55a/logs',
      );
      render(<PipelineDatumViewer />);

      await click(await screen.findByTestId('SidePanel__closeModal'));

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/lineage/Solar-Panel-Data-Sorting/pipelines/montage',
        ),
      );
    });

    it('should route user back to job view on close', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
      );

      render(<JobDatumViewer />);

      await click(await screen.findByTestId('SidePanel__closeModal'));

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/project/Solar-Panel-Data-Sorting/jobs/subjobs',
        ),
      );
    });
  });

  describe('Right Panel', () => {
    it('should render datum details', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
      );

      render(<JobDatumViewer />);

      expect(await screen.findByText('Success')).toBeVisible();

      const runtimeDropDown = await screen.findByText('6 s');
      expect(runtimeDropDown).toBeVisible();

      userEvent.click(runtimeDropDown);

      expect(await screen.findByText('1 s')).toBeVisible();
      expect(await screen.findByText('1 kB')).toBeVisible();
      expect(await screen.findByText('3 s')).toBeVisible();
      expect(await screen.findByText('2 s')).toBeVisible();
      expect(await screen.findByText('2 kB')).toBeVisible();
    });

    it('should render the root key of input spec', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
      );

      render(<JobDatumViewer />);

      const codeSpec = await screen.findByTestId(
        'ConfigFilePreview__codeElement',
      );
      expect(
        await within(codeSpec).findAllByText((node) => node.includes('edges')),
      ).toHaveLength(2); // Allow code element to load
      expect(codeSpec).toMatchSnapshot();
    });

    it('should render N/A when the runtime data is not available', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f14f',
      );

      render(<JobDatumViewer />);

      const runtimeDropDown = await screen.findByText('N/A');
      expect(runtimeDropDown).toBeVisible();

      userEvent.click(runtimeDropDown);

      await screen.findByText('Download'); // Allow dropdown to finish loading
      expect(await screen.findAllByText('N/A')).toHaveLength(4);
    });

    it('should render a skipped datums details', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/1112b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
      );

      render(<JobDatumViewer />);

      expect(await screen.findByText('Skipped')).toBeVisible();
      expect(
        await screen.findByText(
          'This datum has been successfully processed in a previous job, has not changed since then, and therefore, it was skipped in the current job.',
        ),
      ).toBeVisible();
      expect(await screen.findByText('Previous Job')).toBeVisible();
      expect(
        await screen.findByText(
          '2222b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        ),
      ).toBeVisible();
      expect(await screen.findByText('6 s')).toBeVisible();
    });
  });

  describe('Left Panel', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
      );
    });

    it('should load correct empty states for datums list', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/edges/logs',
      );
      render(<JobDatumViewer />);

      await click((await screen.findAllByTestId('JobList__listItem'))[0]);
      expect(
        await screen.findByText('No datums found for this job.'),
      ).toBeVisible();
    });

    it('should format job timestamp correctly', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/edges/logs',
      );
      render(<JobDatumViewer />);
      expect(
        (await screen.findAllByTestId('JobList__listItem'))[0],
      ).toHaveTextContent(
        `${getStandardDate(1614126189)}23b9af7d5d4343219bc8e02ff44cd55a`,
      );
    });

    it('should load datum filters from the url', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155?datumFilters=FAILED,SKIPPED',
      );

      render(<JobDatumViewer />);

      await screen.findByTestId('Filter__FAILEDChip');
      await screen.findByTestId('Filter__SKIPPEDChip');

      const datums = await screen.findAllByTestId('DatumList__listItem');
      const datum = datums[0];
      expect(datum).toHaveTextContent(
        '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
      );
    });

    it('should allow users to set and remove the datum filters', async () => {
      render(<JobDatumViewer />);

      const jobs = await screen.findAllByTestId('JobList__listItem');
      expect(jobs).toHaveLength(4);

      await click(jobs[0]);
      expect(await screen.findAllByTestId('DatumList__listItem')).toHaveLength(
        4,
      );

      await click(jobs[2]);
      expect(await screen.findAllByTestId('DatumList__listItem')).toHaveLength(
        2,
      );

      await click(jobs[0]);
      await click(await screen.findByText('Filter'));
      await click(await screen.findByText('Failed'));

      const datum = await screen.findByTestId('DatumList__listItem');
      expect(datum).toHaveTextContent(
        '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
      );

      await click(jobs[2]);
      expect(
        await screen.findByText('No datums found for this job.'),
      ).toBeVisible();

      await click(jobs[0]);
      await click(await screen.findByTestId('Filter__FAILEDChip'));
      expect(await screen.findAllByTestId('DatumList__listItem')).toHaveLength(
        4,
      );
    });

    it('should sort jobs by status', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
      );

      render(<JobDatumViewer />);
      let jobs = await screen.findAllByTestId('JobList__listItem');
      expect(jobs).toHaveLength(4);

      expect(jobs[0]).toHaveTextContent('23b9af7d5d4343219bc8e02ff44cd55a');
      expect(jobs[1]).toHaveTextContent('33b9af7d5d4343219bc8e02ff44cd55a');
      expect(jobs[2]).toHaveTextContent('7798fhje5d4343219bc8e02ff4acd33a');
      expect(jobs[3]).toHaveTextContent('o90du4js5d4343219bc8e02ff4acd33a');

      await click(await screen.findByText('Filter'));
      await click(await screen.findByText('Job status'));

      jobs = await screen.findAllByTestId('JobList__listItem');
      expect(jobs).toHaveLength(4);
      expect(jobs[0]).toHaveTextContent('33b9af7d5d4343219bc8e02ff44cd55a');
      expect(jobs[1]).toHaveTextContent('o90du4js5d4343219bc8e02ff4acd33a');
      expect(jobs[2]).toHaveTextContent('7798fhje5d4343219bc8e02ff4acd33a');
      expect(jobs[3]).toHaveTextContent('23b9af7d5d4343219bc8e02ff44cd55a');
    });

    describe('Jobs', () => {
      beforeEach(() => {
        window.history.replaceState(
          {},
          '',
          '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
        );
      });

      it('should load jobs and select job from url', async () => {
        render(<JobDatumViewer />);

        expect(
          await screen.findByTestId('BreadCrumbs__base'),
        ).toHaveTextContent('Job: 23b9af7d5d4343219bc8e02ff44cd55a');
        const jobs = await screen.findAllByTestId('JobList__listItem');
        expect(jobs).toHaveLength(4);
        expect(jobs[0]).toHaveClass('selected');
        expect(jobs[0]).toHaveTextContent('23b9af7d5d4343219bc8e02ff44cd55a');

        expect(
          screen.queryByTestId('DatumList__listItem'),
        ).not.toBeInTheDocument();
      });
    });

    describe('Datums', () => {
      beforeEach(() => {
        window.history.replaceState(
          {},
          '',
          '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );
      });

      it('should load datums and select datum from url', async () => {
        render(<JobDatumViewer />);

        const selectedDatum = (
          await screen.findAllByTestId('DatumList__listItem')
        )[0];
        expect(selectedDatum).toHaveClass('selected');
        expect(selectedDatum).toHaveTextContent(
          '0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );
        expect(
          await screen.findByTestId('BreadCrumbs__base'),
        ).toHaveTextContent(
          '.../Datum: 0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );
        expect(
          screen.queryByTestId('JobList__listItem'),
        ).not.toBeInTheDocument();
        await click(selectedDatum);

        expect(
          await screen.findByTestId('BreadCrumbs__base'),
        ).toHaveTextContent(
          '.../Datum: 0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );
      });

      it('should allow users to search for a datum', async () => {
        render(<JobDatumViewer />);

        const search = await screen.findByTestId('DatumList__search');

        expect(
          screen.queryByText('No matching datums found'),
        ).not.toBeInTheDocument();

        await type(
          search,
          '006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f333',
        );
        await screen.findByText('No matching datums found');
        await click(await screen.findByTestId('DatumList__searchClear'));
        expect(search).toHaveTextContent('');
        await type(search, 'werweriuowiejrklwkejrwiepriojw');
        expect(
          screen.getByText('Enter the exact datum ID'),
        ).toBeInTheDocument();
        await click(await screen.findByTestId('DatumList__searchClear'));
        expect(search).toHaveTextContent('');
        await type(
          search,
          '006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f14f',
        );

        const selectedDatum = await screen.findByTestId('DatumList__listItem');
        expect(selectedDatum).toHaveTextContent(
          '006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f14f',
        );
        expect(
          screen.queryByText('No matching datums found'),
        ).not.toBeInTheDocument();
      });

      it('should load the job list if the back button is pressed', async () => {
        render(<JobDatumViewer />);

        await screen.findByTestId('DatumList__list');
        expect(screen.queryByTestId('JobList__list')).not.toBeInTheDocument();
        await click(await screen.findByText('Back'));
        await screen.findByTestId('JobList__list');
        expect(screen.queryByTestId('DatumList__list')).not.toBeInTheDocument();
      });

      it('should allow a user to page through the datum list', async () => {
        window.history.replaceState(
          {},
          '',
          '/project/Data-Cleaning-Process/pipelines/likelihoods/jobs/23b9af7d5d4343219bc8e02ff4acd33a/logs/datum',
        );
        render(<JobDatumViewer />);
        const forwards = (await screen.findAllByTestId('Pager__forward'))[0];
        const backwards = (await screen.findAllByTestId('Pager__backward'))[0];

        expect(
          await screen.findByText('Datums 1 - 50 of 100'),
        ).toBeInTheDocument();
        expect(backwards).toBeDisabled();
        let datums = await screen.findAllByTestId('DatumList__listItem');
        expect(datums[0]).toHaveTextContent(
          '0a00000000000000000000000000000000000000000000000000000000000000',
        );
        expect(datums[49]).toHaveTextContent(
          '49a0000000000000000000000000000000000000000000000000000000000000',
        );
        await click(forwards);
        expect(
          await screen.findByText('Datums 51 - 100 of 100'),
        ).toBeInTheDocument();
        expect(forwards).toBeDisabled();
        datums = await screen.findAllByTestId('DatumList__listItem');
        expect(datums[0]).toHaveTextContent(
          '50a0000000000000000000000000000000000000000000000000000000000000',
        );
        expect(datums[49]).toHaveTextContent(
          '99a0000000000000000000000000000000000000000000000000000000000000',
        );
      });

      it('should calculate total numper of datums for filtered datum list', async () => {
        window.history.replaceState(
          '',
          '',
          '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum',
        );
        render(<JobDatumViewer />);
        expect(
          await screen.findByText('Datums 1 - 4 of 4'),
        ).toBeInTheDocument();

        await click(await screen.findByText('Filter'));
        await click((await screen.findAllByText('Failed'))[0]);
        expect(
          await screen.findByText('Datums 1 - 1 of 1'),
        ).toBeInTheDocument();

        await click(await screen.findByText('Filter'));
        await click((await screen.findAllByText('Success'))[0]);
        expect(
          await screen.findByText('Datums 1 - 3 of 3'),
        ).toBeInTheDocument();
      });

      it('should display loading message for datum info if job is not finshed', async () => {
        window.history.replaceState(
          '',
          '',
          '/project/Solar-Panel-Data-Sorting/jobs/7798fhje5d4343219bc8e02ff4acd33a/pipeline/montage/logs/datum',
        );

        render(<JobDatumViewer />);
        expect(
          await screen.findByTestId('DatumList__processing'),
        ).toHaveTextContent('Processing — datums are being processed.');
      });
    });
  });

  describe('Logs Viewer', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/33b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
      );
    });

    afterEach(() => {
      window.localStorage.removeItem(
        'pachyderm-console-Solar-Panel-Data-Sorting',
      );
    });

    it('should display empty state when there are no logs', async () => {
      render(<MiddleSection />);

      expect(await screen.findAllByTestId('LogRow__checkbox')).toHaveLength(2);

      const filterButton = await screen.findByTestId('DropdownButton__button');
      await click(filterButton);

      const otherOption = screen.getByText('Last 30 Minutes');
      await click(otherOption);
      expect(
        await screen.findByText('No logs found for this time range.'),
      ).toBeInTheDocument();
    });

    it('export options should download and copy selected logs', async () => {
      render(<MiddleSection />);

      const selectAll = await screen.findByTestId('LogsListHeader__select_all');
      const downloadButton = await screen.findByRole('button', {
        name: 'Download selected logs',
      });

      const copyButton = await screen.findByRole('button', {
        name: 'Copy selected logs',
      });

      await screen.findAllByTestId('LogRow__checkbox');

      expect(copyButton).toBeDisabled();
      expect(downloadButton).toBeDisabled();

      await click(selectAll);

      expect(copyButton).toBeEnabled();
      expect(downloadButton).toBeEnabled();

      const selectedLogs = `${getStandardDate(
        1614126189,
      )} started datum task\n${getStandardDate(
        1614126190,
      )} finished datum task`;

      expect(copy).toHaveBeenCalledTimes(0);
      await click(copyButton);
      expect(copy).toHaveBeenCalledTimes(1);
      expect(mockUseClipboardCopy).toHaveBeenLastCalledWith(selectedLogs);

      expect(download).toHaveBeenCalledTimes(0);
      await click(downloadButton);
      expect(download).toHaveBeenCalledTimes(1);
      expect(mockUseDownloadText).toHaveBeenLastCalledWith(
        selectedLogs,
        'montage_logs',
      );
    });

    it('should display logs', async () => {
      render(<MiddleSection />);

      const rows = await screen.findAllByTestId('LogRow__base');
      expect(rows).toHaveLength(2);
      expect(rows[0]).toHaveTextContent(
        `${getStandardDate(1614126189)} started datum task`,
      );
      expect(rows[1]).toHaveTextContent(
        `${getStandardDate(1614126190)} finished datum task`,
      );
    });

    it('should highlight user logs', async () => {
      render(<MiddleSection />);
      await screen.findAllByTestId('LogRow__base');
      expect(screen.queryByTestId('LogRow__user_log')).not.toBeInTheDocument();

      await click(await screen.findByTestId('DropdownButton__button'));
      await click(screen.getByText('Highlight User Logs'));

      expect(await screen.findByTestId('LogRow__user_log')).toBeInTheDocument();
    });

    it('should display raw logs', async () => {
      render(<MiddleSection />);

      await click(await screen.findByTestId('DropdownButton__button'));
      await click(screen.getByText('Raw Logs'));

      const rows = await screen.findAllByTestId('RawLogRow__base');
      expect(rows).toHaveLength(2);
      expect(rows[0]).toHaveTextContent('started datum task');
      expect(rows[1]).toHaveTextContent('finished datum task');
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
      expect(
        await screen.findByTestId('RawLogRow__user_log'),
      ).toBeInTheDocument();
    });

    it('should store logs view preference on refresh', async () => {
      render(<MiddleSection />);

      expect(await screen.findAllByTestId('LogRow__base')).toHaveLength(2);
      expect(screen.queryByTestId('RawLogRow__base')).not.toBeInTheDocument();
      await click(await screen.findByTestId('DropdownButton__button'));
      await click(screen.getByText('Raw Logs'));
      expect(screen.queryByTestId('LogRow__base')).not.toBeInTheDocument();
      expect(await screen.findAllByTestId('RawLogRow__base')).toHaveLength(2);

      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/33b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
      );
      expect(screen.queryByTestId('LogRow__base')).not.toBeInTheDocument();
      expect(await screen.findAllByTestId('RawLogRow__base')).toHaveLength(2);
    });

    it('should scroll to the latest log on first load', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Data-Cleaning-Process/jobs/23b9af7d5d4343219bc8e02ff4acd33a/pipeline/likelihoods/logs',
      );
      render(<MiddleSection />);
      expect(
        (await screen.findAllByTestId('LogRow__checkbox')).length,
      ).toBeGreaterThan(2);

      await screen.findByText(/last message/);
    });

    describe('Job Logs Viewer', () => {
      beforeEach(() => {
        window.history.replaceState(
          {},
          '',
          '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
        );
      });

      it('should show the correct breadcrumbs and title for job logs', async () => {
        render(<MiddleSection />);
        expect(
          await screen.findByTestId('DatumHeaderBreadCrumbs__path'),
        ).toHaveTextContent(
          'Pipeline.../Job: 23b9af7d5d4343219bc8e02ff44cd55a',
        );

        expect(
          await screen.findByTestId('MiddleSection__title'),
        ).toHaveTextContent('Job Logs for23b9af7d5d4343219bc8e02ff44cd55a');
      });

      it('should display all logs for a job', async () => {
        render(<MiddleSection />);
        const rows = await screen.findAllByTestId('LogRow__base');
        expect(rows).toHaveLength(6);
        expect(rows[0]).toHaveTextContent(
          `${getStandardDate(1616533099)} started datum task`,
        );
        expect(rows[5]).toHaveTextContent(
          `${getStandardDate(1616533220)} finished datum task`,
        );
      });
    });

    describe('Datum Logs Viewer', () => {
      beforeEach(() => {
        window.history.replaceState(
          {},
          '',
          '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );
      });

      it('should show the correct breadcrumbs and title for datum logs', async () => {
        render(<MiddleSection />);
        expect(
          await screen.findByTestId('DatumHeaderBreadCrumbs__path'),
        ).toHaveTextContent(
          'Pipeline.../Job.../Datum: 0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );

        expect(
          await screen.findByTestId('MiddleSection__title'),
        ).toHaveTextContent(
          'Datum Logs for0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );

        await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

        const link = screen.getByRole('link', {
          name: 'Job...',
        });
        expect(link).toHaveAttribute(
          'href',
          '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
        );
        await click(link);

        expect(
          await screen.findByTestId('DatumHeaderBreadCrumbs__path'),
        ).toHaveTextContent(
          'Pipeline.../Job: 23b9af7d5d4343219bc8e02ff44cd55a',
        );
      });

      it('should display logs for a given datum', async () => {
        render(<MiddleSection />);
        const rows = await screen.findAllByTestId('LogRow__base');
        expect(rows).toHaveLength(4);
        expect(rows[0]).toHaveTextContent(
          `${getStandardDate(1616533099)} started datum task`,
        );
        expect(rows[3]).toHaveTextContent(
          `${getStandardDate(1616533106)} finished datum task`,
        );
      });

      it('should display message for a skipped datum', async () => {
        window.history.replaceState(
          {},
          '',
          '/project/Solar-Panel-Data-Sorting/jobs/7798fhje5d4343219bc8e02ff4acd33a/pipeline/montage/logs/datum/987654321dbb460a649a72bf1a1147159737c785f622c0c149ff89d7fcb66747',
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
    });

    it('should show logs pager', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/Solar-Panel-Data-Sorting/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
      );
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

  describe('Spout Pipeline', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/lineage/Pipelines-Project/pipelines/spout-pipeline/logs',
      );
    });

    it('should hide the left panel', async () => {
      render(<PipelineDatumViewer />);
      await screen.findByTestId('SidePanel__right');
      expect(screen.queryByTestId('SidePanel__left')).not.toBeInTheDocument();
    });

    it('should display correct header', async () => {
      render(<PipelineDatumViewer />);
      expect(
        await screen.findByTestId('MiddleSection__title'),
      ).toHaveTextContent('Pipeline logs forspout-pipeline');
    });

    it('should show pipeline info in the right panel', async () => {
      render(<PipelineDatumViewer />);

      expect(await screen.findByText('Running')).toBeVisible();
      expect(screen.getByLabelText('Pipeline Type')).toHaveTextContent('Spout');
    });

    it('should display all logs for a spout pipeline', async () => {
      render(<MiddleSection />);
      const rows = await screen.findAllByTestId('LogRow__base');
      expect(rows).toHaveLength(2);
      expect(rows[0]).toHaveTextContent(
        `${getStandardDate(1616533099)} Spout Log`,
      );
      expect(rows[1]).toHaveTextContent(
        `${getStandardDate(1616533100)} Spout Log 2`,
      );
    });
  });

  describe('Service Pipeline', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/lineage/Pipelines-Project/pipelines/service-pipeline/jobs/5940382d5d4343219bc8e02ff44cd55a/logs',
      );
    });

    it('should hide the datum panel when clicking on a job', async () => {
      render(<PipelineDatumViewer />);
      await screen.findByTestId('JobList__list');
      await click((await screen.findAllByTestId('JobList__listItem'))[1]);
      expect(screen.queryByTestId('DatumList__list')).not.toBeInTheDocument();
      await screen.findByTestId('JobList__list');
    });

    it('should hide the datum filter options in the right panel', async () => {
      render(<PipelineDatumViewer />);

      await click(await screen.findByText('Filter'));
      await screen.findByText('Sort Jobs By');
      expect(
        screen.queryByText('Filter Datums by Status'),
      ).not.toBeInTheDocument();
    });

    it('should display correct header', async () => {
      render(<PipelineDatumViewer />);
      expect(
        await screen.findByTestId('MiddleSection__title'),
      ).toHaveTextContent('Job Logs for5940382d5d4343219bc8e02ff44cd55a');
    });

    it('should show pipeline info in the right panel', async () => {
      render(<PipelineDatumViewer />);

      expect(await screen.findByText('Running')).toBeVisible();
      expect(screen.getByLabelText('Pipeline Type')).toHaveTextContent(
        'Service',
      );
    });

    it('should display correct logs for a service pipeline', async () => {
      render(<PipelineDatumViewer />);

      expect(await screen.findByTestId('LogRow__base')).toHaveTextContent(
        `${getStandardDate(1616533098)} service log running`,
      );
      await click((await screen.findAllByTestId('JobList__listItem'))[1]);
      expect(await screen.findByTestId('LogRow__base')).toHaveTextContent(
        `${getStandardDate(1616533098)} service log complete`,
      );
    });
  });
});
