import {render, waitFor, within} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {format, fromUnixTime} from 'date-fns';
import React from 'react';

import {useClipboardCopy} from '@dash-frontend/../components/src';
import useDownloadText from '@dash-frontend/hooks/useDownloadText';
import {withContextProviders, click, type} from '@dash-frontend/testHelpers';

import {default as MiddleSectionComponent} from '../components/MiddleSection';
import {LOGS_DATE_FORMAT} from '../components/MiddleSection/components/LogsViewer/constants/logsViewersConstants';
import {JOB_DATE_FORMAT} from '../constants/DatumViewer';
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
        '/lineage/1/pipelines/montage/jobs/23b9af7d5d4343219bc8e02ff44cd55a/logs',
      );
      const {findByTestId} = render(<PipelineDatumViewer />);

      await click(await findByTestId('SidePanel__closeModal'));

      await waitFor(() =>
        expect(window.location.pathname).toBe('/lineage/1/pipelines/montage'),
      );
    });

    it('should route user back to job view on close', async () => {
      window.history.replaceState(
        {},
        '',
        '/lineage/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
      );

      const {findByTestId} = render(<JobDatumViewer />);

      await click(await findByTestId('SidePanel__closeModal'));

      await waitFor(() =>
        expect(window.location.pathname).toBe(
          '/lineage/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/montage',
        ),
      );
    });
  });

  describe('Right Panel', () => {
    it('should render datum details', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
      );

      const {findByText} = render(<JobDatumViewer />);

      expect(await findByText('Success')).toBeVisible();

      const runtimeDropDown = await findByText('6.15 seconds');
      expect(runtimeDropDown).toBeVisible();

      userEvent.click(runtimeDropDown);

      expect(await findByText('1.12 seconds')).toBeVisible();
      expect(await findByText('1 kB')).toBeVisible();
      expect(await findByText('3 seconds')).toBeVisible();
      expect(await findByText('2.02 seconds')).toBeVisible();
      expect(await findByText('2 kB')).toBeVisible();
    });

    it('should render the root key of input spec', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
      );

      const {getByTestId} = render(<JobDatumViewer />);

      const yamlSpec = getByTestId('ConfigFilePreview__codeElement');
      expect(await within(yamlSpec).findAllByText('edges')).toHaveLength(2); // Allow code element to load
      expect(yamlSpec).toMatchSnapshot();
    });

    it('should render N/A when the runtime data is not available', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f14f',
      );

      const {findByText, findAllByText} = render(<JobDatumViewer />);

      const runtimeDropDown = await findByText('N/A');
      expect(runtimeDropDown).toBeVisible();

      userEvent.click(runtimeDropDown);

      await findByText('Download'); // Allow dropdown to finish loading
      expect(await findAllByText('N/A')).toHaveLength(4);
    });

    it('should render a skipped datums details', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/1112b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
      );

      const {findByText} = render(<JobDatumViewer />);

      expect(await findByText('Skipped')).toBeVisible();
      expect(
        await findByText(
          'This datum has been successfully processed in a previous job, has not changed since then, and therefore, it was skipped in the current job.',
        ),
      ).toBeVisible();
      expect(await findByText('Previous Job')).toBeVisible();
      expect(
        await findByText(
          '2222b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        ),
      ).toBeVisible();
      expect(await findByText('6.15 seconds')).toBeVisible();
    });
  });

  describe('Left Panel', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
      );
    });

    it('should load correct empty states for datums list', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/edges/logs',
      );
      const {findByText, findAllByTestId} = render(<JobDatumViewer />);

      await click((await findAllByTestId('JobList__listItem'))[0]);
      expect(await findByText('No datums found for this job.')).toBeVisible();
    });

    it('should format job timestamp correctly', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/edges/logs',
      );
      const {findAllByTestId} = render(<JobDatumViewer />);
      expect(
        (await findAllByTestId('JobList__listItem'))[0].textContent,
      ).toEqual(
        `${format(
          fromUnixTime(1614126189),
          JOB_DATE_FORMAT,
        )}23b9af7d5d4343219bc8e02ff44cd55a`,
      );
    });

    it('should load datum filters from the url', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155?view=eyJkYXR1bUZpbHRlcnMiOlsiU1RBUlRJTkciLCJTS0lQUEVEIl19',
      );

      const {findAllByTestId, findByTestId} = render(<JobDatumViewer />);

      await findByTestId('Filter__STARTINGChip');
      await findByTestId('Filter__SKIPPEDChip');

      const datums = await findAllByTestId('DatumList__listItem');
      const datum = datums[0];
      expect(datum.textContent).toEqual(
        '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
      );
    });

    it('should allow users to set and remove the datum filters', async () => {
      const {findByText, findAllByTestId, findByTestId} = render(
        <JobDatumViewer />,
      );

      const jobs = await findAllByTestId('JobList__listItem');
      expect(jobs.length).toEqual(4);

      await click(jobs[0]);
      expect((await findAllByTestId('DatumList__listItem')).length).toEqual(4);

      await click(jobs[2]);
      expect((await findAllByTestId('DatumList__listItem')).length).toEqual(2);

      await click(jobs[0]);
      await click(await findByText('Filter'));
      await click(await findByText('Starting'));

      const datum = await findByTestId('DatumList__listItem');
      expect(datum.textContent).toEqual(
        '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
      );

      await click(jobs[2]);
      expect(await findByText('No datums found for this job.')).toBeVisible();

      await click(jobs[0]);
      await click(await findByTestId('Filter__STARTINGChip'));
      expect((await findAllByTestId('DatumList__listItem')).length).toEqual(4);
    });

    it('should sort jobs by status', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
      );

      const {findByText, findAllByTestId} = render(<JobDatumViewer />);
      let jobs = await findAllByTestId('JobList__listItem');
      expect(jobs.length).toEqual(4);

      expect(jobs[0]).toHaveTextContent('23b9af7d5d4343219bc8e02ff44cd55a');
      expect(jobs[1]).toHaveTextContent('33b9af7d5d4343219bc8e02ff44cd55a');
      expect(jobs[2]).toHaveTextContent('7798fhje5d4343219bc8e02ff4acd33a');
      expect(jobs[3]).toHaveTextContent('o90du4js5d4343219bc8e02ff4acd33a');

      await click(await findByText('Filter'));
      await click(await findByText('Job status'));

      jobs = await findAllByTestId('JobList__listItem');
      expect(jobs.length).toEqual(4);
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
          '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
        );
      });

      it('should load jobs and select job from url', async () => {
        const {findAllByTestId, findByTestId, queryAllByTestId} = render(
          <JobDatumViewer />,
        );

        expect(await findByTestId('BreadCrumbs__base')).toHaveTextContent(
          'Job: 23b9af7d5d4343219bc8e02ff44cd55a',
        );
        const jobs = await findAllByTestId('JobList__listItem');
        expect(jobs.length).toEqual(4);
        expect(jobs[0]).toHaveClass('selected');
        expect(jobs[0]).toHaveTextContent('23b9af7d5d4343219bc8e02ff44cd55a');

        expect(queryAllByTestId('DatumList__listItem').length).toEqual(0);
      });
    });

    describe('Datums', () => {
      beforeEach(() => {
        window.history.replaceState(
          {},
          '',
          '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );
      });

      it('should load datums and select datum from url', async () => {
        const {findAllByTestId, queryAllByTestId, findByTestId} = render(
          <JobDatumViewer />,
        );

        const selectedDatum = (await findAllByTestId('DatumList__listItem'))[0];
        expect(selectedDatum).toHaveClass('selected');
        expect(selectedDatum.textContent).toEqual(
          '0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );
        expect((await findByTestId('BreadCrumbs__base')).textContent).toEqual(
          '.../Datum: 0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );
        expect(queryAllByTestId('JobList__listItem').length).toEqual(0);
        await click(selectedDatum);

        expect((await findByTestId('BreadCrumbs__base')).textContent).toEqual(
          '.../Datum: 0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );
      });

      it('should allow users to search for a datum', async () => {
        const {findByTestId, queryByText, findByText} = render(
          <JobDatumViewer />,
        );

        const search = await findByTestId('DatumList__search');

        expect(queryByText('No matching datums found')).not.toBeInTheDocument();

        await type(
          search,
          '006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f333',
        );
        await findByText('No matching datums found');
        await click(await findByTestId('DatumList__searchClear'));
        expect(search.textContent).toEqual('');
        await type(
          search,
          '006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f14f',
        );

        const selectedDatum = await findByTestId('DatumList__listItem');
        expect(selectedDatum.textContent).toEqual(
          '006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f14f',
        );
        expect(queryByText('No matching datums found')).not.toBeInTheDocument();
      });

      it('should load the job list if the back button is pressed', async () => {
        const {findByTestId, queryByTestId, findByText} = render(
          <JobDatumViewer />,
        );

        await findByTestId('DatumList__list');
        expect(queryByTestId('JobList__list')).not.toBeInTheDocument();
        await click(await findByText('Back'));
        await findByTestId('JobList__list');
        expect(queryByTestId('DatumList__list')).not.toBeInTheDocument();
      });
    });
  });

  describe('Logs Viewer', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        '/project/1/jobs/33b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
      );
    });

    afterEach(() => {
      window.localStorage.removeItem('pachyderm-console-1');
    });

    it('should display empty state when there are no logs', async () => {
      const {findByTestId, getByText, findAllByTestId, findByText} = render(
        <MiddleSection />,
      );

      expect(await findAllByTestId('LogRow__checkbox')).toHaveLength(2);

      const filterButton = await findByTestId('DropdownButton__button');
      await click(filterButton);

      const otherOption = getByText('Last 30 Minutes');
      await click(otherOption);
      expect(
        await findByText('No logs found for this time range.'),
      ).toBeInTheDocument();
    });

    it('export options should download and copy selected logs', async () => {
      const {findByTestId, findAllByTestId, findByRole} = render(
        <MiddleSection />,
      );

      const selectAll = await findByTestId('LogsListHeader__select_all');
      const downloadButton = await findByRole('button', {
        name: 'Download selected logs',
      });

      const copyButton = await findByRole('button', {
        name: 'Copy selected logs',
      });

      await findAllByTestId('LogRow__checkbox');

      expect(copyButton).toBeDisabled();
      expect(downloadButton).toBeDisabled();

      await click(selectAll);

      expect(copyButton).not.toBeDisabled();
      expect(downloadButton).not.toBeDisabled();

      const selectedLogs = `${format(
        fromUnixTime(1614126189),
        LOGS_DATE_FORMAT,
      )} started datum task\n${format(
        fromUnixTime(1614126190),
        LOGS_DATE_FORMAT,
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
      const {findAllByTestId} = render(<MiddleSection />);

      const rows = await findAllByTestId('LogRow__base');
      expect(rows).toHaveLength(2);
      expect(rows[0].textContent).toEqual(
        `${format(
          fromUnixTime(1614126189),
          LOGS_DATE_FORMAT,
        )} started datum task`,
      );
      expect(rows[1].textContent).toEqual(
        `${format(
          fromUnixTime(1614126190),
          LOGS_DATE_FORMAT,
        )} finished datum task`,
      );
    });

    it('should highlight user logs', async () => {
      const {queryAllByTestId, findAllByTestId, findByTestId, getByText} =
        render(<MiddleSection />);
      expect(queryAllByTestId('LogRow__user_log')).toHaveLength(0);

      await click(await findByTestId('DropdownButton__button'));
      await click(getByText('Highlight User Logs'));

      expect(await findAllByTestId('LogRow__user_log')).toHaveLength(1);
    });

    it('should display raw logs', async () => {
      const {findAllByTestId, findByTestId, getByText} = render(
        <MiddleSection />,
      );

      await click(await findByTestId('DropdownButton__button'));
      await click(getByText('Raw Logs'));

      const rows = await findAllByTestId('RawLogRow__base');
      expect(rows).toHaveLength(2);
      expect(rows[0].textContent).toEqual('started datum task');
      expect(rows[1].textContent).toEqual('finished datum task');
    });

    it('should highlight raw user logs', async () => {
      const {queryAllByTestId, findAllByTestId, findByTestId, getByText} =
        render(<MiddleSection />);

      await click(await findByTestId('DropdownButton__button'));
      await click(getByText('Raw Logs'));
      expect(queryAllByTestId('RawLogRow__user_log')).toHaveLength(0);
      await click(getByText('Highlight User Logs'));
      expect(await findAllByTestId('RawLogRow__user_log')).toHaveLength(1);
    });

    it('should store logs view preference on refresh', async () => {
      const {findAllByTestId, queryAllByTestId, findByTestId, getByText} =
        render(<MiddleSection />);

      expect(await findAllByTestId('LogRow__base')).toHaveLength(2);
      expect(queryAllByTestId('RawLogRow__base')).toHaveLength(0);
      await click(await findByTestId('DropdownButton__button'));
      await click(getByText('Raw Logs'));
      expect(queryAllByTestId('LogRow__base')).toHaveLength(0);
      expect(await findAllByTestId('RawLogRow__base')).toHaveLength(2);

      window.history.replaceState(
        {},
        '',
        '/project/1/jobs/33b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
      );
      expect(queryAllByTestId('LogRow__base')).toHaveLength(0);
      expect(await findAllByTestId('RawLogRow__base')).toHaveLength(2);
    });

    it('should scroll to the latest log on first load', async () => {
      window.history.replaceState(
        {},
        '',
        '/project/2/jobs/23b9af7d5d4343219bc8e02ff4acd33a/pipeline/likelihoods/logs',
      );
      const {queryByText, findAllByTestId} = render(<MiddleSection />);
      expect(
        (await findAllByTestId('LogRow__checkbox')).length,
      ).toBeGreaterThan(2);

      await waitFor(() =>
        expect(queryByText(/last message/)).toBeInTheDocument(),
      );
    });

    describe('Job Logs Viewer', () => {
      beforeEach(() => {
        window.history.replaceState(
          {},
          '',
          '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs',
        );
      });

      it('should show the correct breadcrumbs and title for job logs', async () => {
        const {findByTestId} = render(<MiddleSection />);
        expect(
          (await findByTestId('DatumHeaderBreadCrumbs__path')).textContent,
        ).toEqual('Pipeline.../Job: 23b9af7d5d4343219bc8e02ff44cd55a');

        expect(
          (await findByTestId('MiddleSection__title')).textContent,
        ).toEqual('Job Logs for23b9af7d5d4343219bc8e02ff44cd55a');
      });

      it('should display all logs for a job', async () => {
        const {findAllByTestId} = render(<MiddleSection />);
        const rows = await findAllByTestId('LogRow__base');
        expect(rows).toHaveLength(6);
        expect(rows[0].textContent).toEqual(
          `${format(
            fromUnixTime(1616533099),
            LOGS_DATE_FORMAT,
          )} started datum task`,
        );
        expect(rows[5].textContent).toEqual(
          `${format(
            fromUnixTime(1616533220),
            LOGS_DATE_FORMAT,
          )} finished datum task`,
        );
      });
    });

    describe('Datum Logs Viewer', () => {
      beforeEach(() => {
        window.history.replaceState(
          {},
          '',
          '/project/1/jobs/23b9af7d5d4343219bc8e02ff44cd55a/pipeline/montage/logs/datum/0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );
      });

      it('should show the correct breadcrumbs and title for datum logs', async () => {
        const {findByTestId, findByText} = render(<MiddleSection />);
        expect(
          (await findByTestId('DatumHeaderBreadCrumbs__path')).textContent,
        ).toEqual(
          'Pipeline.../Job.../Datum: 0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );

        expect(
          (await findByTestId('MiddleSection__title')).textContent,
        ).toEqual(
          'Datum Logs for0752b20131461a629431125793336672cdf30fff4a01406021603bbc98b4255d',
        );

        await click(await findByText('Job...'));

        expect(
          (await findByTestId('DatumHeaderBreadCrumbs__path')).textContent,
        ).toEqual('Pipeline.../Job: 23b9af7d5d4343219bc8e02ff44cd55a');
      });

      it('should display logs for a given datum', async () => {
        const {findAllByTestId} = render(<MiddleSection />);
        const rows = await findAllByTestId('LogRow__base');
        expect(rows).toHaveLength(4);
        expect(rows[0].textContent).toEqual(
          `${format(
            fromUnixTime(1616533099),
            LOGS_DATE_FORMAT,
          )} started datum task`,
        );
        expect(rows[3].textContent).toEqual(
          `${format(
            fromUnixTime(1616533106),
            LOGS_DATE_FORMAT,
          )} finished datum task`,
        );
      });

      it('should display message for a skipped datum', async () => {
        window.history.replaceState(
          {},
          '',
          '/project/1/jobs/7798fhje5d4343219bc8e02ff4acd33a/pipeline/montage/logs/datum/987654321dbb460a649a72bf1a1147159737c785f622c0c149ff89d7fcb66747',
        );
        const {findByText} = render(<MiddleSection />);
        expect(
          await findByText(
            'This datum has been successfully processed in a previous job.',
          ),
        ).toBeInTheDocument();
      });
    });
  });
});
