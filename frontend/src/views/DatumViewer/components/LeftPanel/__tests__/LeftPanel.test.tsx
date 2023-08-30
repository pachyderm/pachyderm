import {
  DatumFilter,
  DatumState,
  JobState,
  mockDatumSearchQuery,
  mockDatumsQuery,
  mockJobQuery,
  mockJobsQuery,
} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import isEqual from 'lodash/isEqual';
import {setupServer} from 'msw/node';
import React from 'react';

import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {
  mockEmptyGetAuthorize,
  mockGetVersionInfo,
  buildDatum,
  buildJob,
  MOCK_EMPTY_DATUMS,
  mockEmptyDatumsQuery,
  mockGetMontagePipeline,
  mockGetMontageJob_5C,
  mockGetMontageJobs,
  mockGetJob5CDatums,
  JOB_5C_DATUM_05,
  mockGetJob5CDatum05,
  generatePagingDatums,
} from '@dash-frontend/mocks';
import {click, type, withContextProviders} from '@dash-frontend/testHelpers';

import LeftPanelComponent from '../LeftPanel';

const basePath =
  '/project/default/jobs/5c1aa9bc87dd411ba5a1be0c80a3ebc2/pipeline/montage/logs';

describe('Datum Viewer Left Panel', () => {
  const server = setupServer();

  const LeftPanel = withContextProviders(() => {
    return <LeftPanelComponent />;
  });

  beforeAll(() => server.listen());

  beforeEach(() => {
    window.history.replaceState({}, '', basePath);
    server.resetHandlers();
    server.use(mockEmptyGetAuthorize());
    server.use(mockGetVersionInfo());
    server.use(mockGetMontagePipeline());
    server.use(mockGetMontageJob_5C());
    server.use(mockGetMontageJobs());
    server.use(mockGetJob5CDatums());
    server.use(mockGetJob5CDatum05());
  });

  afterAll(() => server.close());

  describe('Jobs', () => {
    it('should format job timestamp correctly', async () => {
      render(<LeftPanel />);

      expect(
        (await screen.findAllByTestId('JobList__listItem'))[0],
      ).toHaveTextContent(
        `${getStandardDate(1690899649)}1dc67e479f03498badcc6180be4ee6ce`,
      );
    });

    it('should load jobs and select job from url', async () => {
      render(<LeftPanel />);

      expect(await screen.findByTestId('BreadCrumbs__base')).toHaveTextContent(
        'Job: 5c1aa9bc87dd411ba5a1be0c80a3ebc2',
      );
      const jobs = await screen.findAllByTestId('JobList__listItem');
      expect(jobs).toHaveLength(4);
      expect(jobs[2]).toHaveClass('selected');
      expect(jobs[2]).toHaveTextContent('5c1aa9bc87dd411ba5a1be0c80a3ebc2');

      expect(
        screen.queryByTestId('DatumList__listItem'),
      ).not.toBeInTheDocument();
    });

    it('should sort jobs by status', async () => {
      server.use(
        mockJobsQuery((_req, res, ctx) => {
          return res(
            ctx.data({
              jobs: {
                items: [
                  buildJob({
                    id: '23b9af7d5d4343219bc8e02ff44cd55a',
                    state: JobState.JOB_SUCCESS,
                    createdAt: 1616533099,
                    pipelineName: 'montage',
                  }),
                  buildJob({
                    id: '33b9af7d5d4343219bc8e02ff44cd55a',
                    state: JobState.JOB_FAILURE,
                    createdAt: 1614126190,
                    pipelineName: 'montage',
                  }),
                  buildJob({
                    id: '7798fhje5d4343219bc8e02ff4acd33a',
                    state: JobState.JOB_FINISHING,
                    createdAt: 1614125000,
                    pipelineName: 'montage',
                  }),
                  buildJob({
                    id: 'o90du4js5d4343219bc8e02ff4acd33a',
                    state: JobState.JOB_KILLED,
                    createdAt: 1614123000,
                    pipelineName: 'montage',
                  }),
                ],
                cursor: null,
                hasNextPage: false,
                __typename: 'PageableJob',
              },
            }),
          );
        }),
      );

      render(<LeftPanel />);
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
  });

  describe('Datums', () => {
    beforeEach(() => {
      window.history.replaceState(
        {},
        '',
        `${basePath}/datum/05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7`,
      );
    });

    it('should load correct empty states for datums list', async () => {
      window.history.replaceState({}, '', basePath);
      server.use(mockEmptyDatumsQuery());

      render(<LeftPanel />);

      await click((await screen.findAllByTestId('JobList__listItem'))[0]);
      expect(
        await screen.findByText('No datums found for this job.'),
      ).toBeVisible();
    });

    it('should load datums and select datum from url', async () => {
      render(<LeftPanel />);
      const selectedDatum = (
        await screen.findAllByTestId('DatumList__listItem')
      )[0];
      expect(selectedDatum).toHaveClass('selected');
      expect(selectedDatum).toHaveTextContent(
        '05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7',
      );
      expect(await screen.findByTestId('BreadCrumbs__base')).toHaveTextContent(
        '.../Datum: 05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7',
      );
      expect(screen.queryByTestId('JobList__listItem')).not.toBeInTheDocument();
      await click(selectedDatum);
      expect(await screen.findByTestId('BreadCrumbs__base')).toHaveTextContent(
        '.../Datum: 05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7',
      );
    });

    it('should load the job list if the back button is pressed', async () => {
      render(<LeftPanel />);

      await screen.findByTestId('DatumList__list');
      expect(screen.queryByTestId('JobList__list')).not.toBeInTheDocument();
      await click(await screen.findByText('Back'));
      await screen.findByTestId('JobList__list');
      expect(screen.queryByTestId('DatumList__list')).not.toBeInTheDocument();
    });

    it('should load datum filters from the url', async () => {
      window.history.replaceState(
        {},
        '',
        `${basePath}/datum/01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155?datumFilters=FAILED,SKIPPED`,
      );
      server.use(
        mockDatumsQuery((req, res, ctx) => {
          if (
            isEqual(req.variables.args.filter, [
              DatumFilter.FAILED,
              DatumFilter.SKIPPED,
            ])
          ) {
            return res(
              ctx.data({
                datums: {
                  items: [
                    buildDatum({
                      id: '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
                      requestedJobId: '23b9af7d5d4343219bc8e02ff44cd55a',
                      state: DatumState.FAILED,
                    }),
                  ],
                  cursor: null,
                  hasNextPage: false,
                  __typename: 'PageableDatum',
                },
              }),
            );
          }
          return res(ctx.data(MOCK_EMPTY_DATUMS));
        }),
      );

      render(<LeftPanel />);

      await screen.findByTestId('Filter__FAILEDChip');
      await screen.findByTestId('Filter__SKIPPEDChip');

      const datums = await screen.findAllByTestId('DatumList__listItem');
      expect(datums[0]).toHaveTextContent(
        '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
      );
    });

    it('should allow users to set and remove the datum filters', async () => {
      window.history.replaceState(
        {},
        '',
        `${basePath}/datum/01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155`,
      );

      server.use(
        mockDatumsQuery((req, res, ctx) => {
          if (isEqual(req.variables.args.filter, [DatumFilter.FAILED])) {
            return res(
              ctx.data({
                datums: {
                  items: [
                    buildDatum({
                      id: '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
                      requestedJobId: '23b9af7d5d4343219bc8e02ff44cd55a',
                      state: DatumState.FAILED,
                    }),
                  ],
                  cursor: null,
                  hasNextPage: false,
                  __typename: 'PageableDatum',
                },
              }),
            );
          }
          return res(
            ctx.data({
              datums: {
                items: [
                  buildDatum({
                    id: '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
                    requestedJobId: '23b9af7d5d4343219bc8e02ff44cd55a',
                    state: DatumState.FAILED,
                  }),

                  buildDatum({
                    id: '006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f14f',
                    requestedJobId: '23b9af7d5d4343219bc8e02ff44cd55a',
                    state: DatumState.SUCCESS,
                  }),
                ],
                cursor: null,
                hasNextPage: false,
                __typename: 'PageableDatum',
              },
            }),
          );
        }),
      );

      render(<LeftPanel />);

      expect(await screen.findAllByTestId('DatumList__listItem')).toHaveLength(
        2,
      );

      await click(await screen.findByText('Filter'));
      await click(await screen.findByText('Failed'));

      const datum = await screen.findByTestId('DatumList__listItem');
      expect(datum).toHaveTextContent(
        '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
      );

      await click(await screen.findByTestId('Filter__FAILEDChip'));
      expect(await screen.findAllByTestId('DatumList__listItem')).toHaveLength(
        2,
      );
    });

    it('should allow users to search for a datum', async () => {
      server.use(
        mockDatumSearchQuery((req, res, ctx) => {
          if (
            req.variables.args.id ===
            '05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7'
          ) {
            return res(ctx.data({datumSearch: JOB_5C_DATUM_05}));
          }
          return res(ctx.data({datumSearch: null}));
        }),
      );

      render(<LeftPanel />);

      const search = await screen.findByTestId('DatumList__search');

      expect(
        screen.queryByText('No matching datums found'),
      ).not.toBeInTheDocument();

      await type(
        search,
        '05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e3929XYZ',
      );
      await screen.findByText('No matching datums found');
      await click(await screen.findByTestId('DatumList__searchClear'));
      expect(search).toHaveTextContent('');
      await type(search, 'werweriuowiejrklwkejrwiepriojw');
      expect(screen.getByText('Enter the exact datum ID')).toBeInTheDocument();
      await click(await screen.findByTestId('DatumList__searchClear'));
      expect(search).toHaveTextContent('');
      await type(
        search,
        '05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7',
      );

      const selectedDatum = await screen.findByTestId('DatumList__listItem');
      expect(selectedDatum).toHaveTextContent(
        '05b864850d01075385e7872e7955fbf710d0e4af0bd73dcf232034a2e39295a7',
      );
      expect(
        screen.queryByText('No matching datums found'),
      ).not.toBeInTheDocument();
    });

    it('should allow a user to page through the datum list', async () => {
      server.use(
        mockJobQuery((_req, res, ctx) => {
          return res(
            ctx.data({
              job: buildJob({
                id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
                state: JobState.JOB_SUCCESS,
                pipelineName: 'montage',
                finishedAt: 1690899618,
                dataProcessed: 100,
                dataSkipped: 0,
                dataFailed: 0,
                dataTotal: 100,
                dataRecovered: 0,
              }),
            }),
          );
        }),
      );

      const datums = generatePagingDatums({n: 100});

      server.use(
        mockDatumsQuery((req, res, ctx) => {
          const {cursor, limit} = req.variables.args;

          if (limit === 50 && !cursor) {
            return res(
              ctx.data({
                datums: {
                  items: datums.slice(0, 50),
                  cursor: datums[49].id,
                  hasNextPage: true,
                },
              }),
            );
          }

          if (limit === 50 && cursor === datums[49].id) {
            return res(
              ctx.data({
                datums: {
                  items: datums.slice(50),
                  cursor: null,
                  hasNextPage: null,
                },
              }),
            );
          }

          return res(
            ctx.data({
              datums: {
                items: datums,
                cursor: null,
                hasNextPage: null,
              },
            }),
          );
        }),
      );

      render(<LeftPanel />);
      const forwards = (await screen.findAllByTestId('Pager__forward'))[0];
      const backwards = (await screen.findAllByTestId('Pager__backward'))[0];

      expect(
        await screen.findByText('Datums 1 - 50 of 100'),
      ).toBeInTheDocument();
      expect(backwards).toBeDisabled();
      let foundDatums = await screen.findAllByTestId('DatumList__listItem');
      expect(foundDatums[0]).toHaveTextContent(
        '0a00000000000000000000000000000000000000000000000000000000000000',
      );
      expect(foundDatums[49]).toHaveTextContent(
        '49a0000000000000000000000000000000000000000000000000000000000000',
      );
      await click(forwards);
      expect(
        await screen.findByText('Datums 51 - 100 of 100'),
      ).toBeInTheDocument();
      expect(forwards).toBeDisabled();
      foundDatums = await screen.findAllByTestId('DatumList__listItem');
      expect(foundDatums[0]).toHaveTextContent(
        '50a0000000000000000000000000000000000000000000000000000000000000',
      );
      expect(foundDatums[49]).toHaveTextContent(
        '99a0000000000000000000000000000000000000000000000000000000000000',
      );
    });

    it('should calculate total number of datums for filtered datum list', async () => {
      server.use(
        mockJobQuery((_req, res, ctx) => {
          return res(
            ctx.data({
              job: buildJob({
                id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
                state: JobState.JOB_SUCCESS,
                pipelineName: 'montage',
                finishedAt: 1690899618,
                dataProcessed: 2,
                dataRecovered: 0,
                dataSkipped: 1,
                dataFailed: 1,
                dataTotal: 4,
              }),
            }),
          );
        }),
      );

      render(<LeftPanel />);
      expect(await screen.findByText('Datums 1 - 4 of 4')).toBeInTheDocument();

      await click(await screen.findByText('Filter'));
      await click((await screen.findAllByText('Failed'))[0]);
      expect(await screen.findByText('Datums 1 - 1 of 1')).toBeInTheDocument();

      await click(await screen.findByText('Filter'));
      await click((await screen.findAllByText('Success'))[0]);
      expect(await screen.findByText('Datums 1 - 3 of 3')).toBeInTheDocument();
    });

    it('should display loading message for datum info if job is not finshed', async () => {
      server.use(
        mockJobQuery((_req, res, ctx) => {
          return res(
            ctx.data({
              job: buildJob({
                id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
                state: JobState.JOB_RUNNING,
                pipelineName: 'montage',
                finishedAt: null,
              }),
            }),
          );
        }),
      );
      render(<LeftPanel />);
      expect(
        await screen.findByTestId('DatumList__processing'),
      ).toHaveTextContent('Processing — datums are being processed.');
    });
  });
});
