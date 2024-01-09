import {
  render,
  screen,
  waitForElementToBeRemoved,
  within,
} from '@testing-library/react';
import isEqual from 'lodash/isEqual';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  DatumInfo,
  DatumState,
  InspectJobRequest,
  JobInfo,
  JobState,
  ListDatumRequest,
  ListJobRequest,
} from '@dash-frontend/api/pps';
import {getISOStringFromUnix} from '@dash-frontend/lib/dateTime';
import {
  generatePagingDatums,
  mockGetEnterpriseInfoInactive,
  mockGetVersionInfo,
  mockGetMontagePipeline,
  mockGetJob5CDatum05,
  mockGetMontageJobs,
  buildJob,
  mockGetJob5CDatums,
  mockEmptyDatums,
  buildDatum,
  MONTAGE_JOB_INFO_5C,
  mockEmptyGetAuthorize,
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
    server.use(
      rest.post<InspectJobRequest, Empty, JobInfo>(
        '/api/pps_v2.API/InspectJob',
        async (req, res, ctx) => {
          const body = await req.json();
          if (body?.job?.pipeline?.project?.name === 'default') {
            return res(ctx.json(MONTAGE_JOB_INFO_5C));
          }
        },
      ),
    );
    server.use(mockGetMontageJobs());
    server.use(mockGetJob5CDatums());
    server.use(mockGetEnterpriseInfoInactive());
  });

  afterAll(() => server.close());

  describe('Jobs', () => {
    it('should format job timestamp correctly', async () => {
      render(<LeftPanel />);
      await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));
      const view = screen.getByTestId('JobList__list');
      expect(within(view).getAllByText(/aug 1, 2023; 14:20/i)).toHaveLength(2);
      expect(
        within(view).getByText(/5c1aa9bc87dd411ba5a1be0c80a3ebc2/i),
      ).toBeInTheDocument();
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
        rest.post<ListJobRequest, Empty, JobInfo[]>(
          '/api/pps_v2.API/ListJob',
          async (req, res, ctx) => {
            const body = await req.json();
            if (body?.projects?.[0]?.name === 'default') {
              return res(
                ctx.json([
                  buildJob({
                    job: {
                      id: '23b9af7d5d4343219bc8e02ff44cd55a',
                      pipeline: {
                        name: 'montage',
                      },
                    },
                    state: JobState.JOB_SUCCESS,
                    created: getISOStringFromUnix(1616533099),
                  }),
                  buildJob({
                    job: {
                      id: '33b9af7d5d4343219bc8e02ff44cd55a',
                      pipeline: {
                        name: 'montage',
                      },
                    },
                    state: JobState.JOB_FAILURE,
                    created: getISOStringFromUnix(1614126190),
                  }),
                  buildJob({
                    job: {
                      id: '7798fhje5d4343219bc8e02ff4acd33a',
                      pipeline: {
                        name: 'montage',
                      },
                    },
                    state: JobState.JOB_FINISHING,
                    created: getISOStringFromUnix(1614125000),
                  }),
                  buildJob({
                    job: {
                      id: 'o90du4js5d4343219bc8e02ff4acd33a',
                      pipeline: {
                        name: 'montage',
                      },
                    },
                    state: JobState.JOB_KILLED,
                    created: getISOStringFromUnix(1614123000),
                  }),
                ]),
              );
            }
          },
        ),
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
      server.use(mockEmptyDatums());

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
        rest.post<ListDatumRequest, Empty, DatumInfo[]>(
          '/api/pps_v2.API/ListDatum',
          async (req, res, ctx) => {
            const jsonReq = await req.json();
            if (
              isEqual(jsonReq.filter.state, [
                DatumState.FAILED,
                DatumState.SKIPPED,
              ])
            ) {
              return res(
                ctx.json([
                  buildDatum({
                    datum: {
                      id: '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
                    },
                    state: DatumState.FAILED,
                  }),
                ]),
              );
            }
            return res(ctx.json([]));
          },
        ),
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
        rest.post<ListDatumRequest, Empty, DatumInfo[]>(
          '/api/pps_v2.API/ListDatum',
          async (req, res, ctx) => {
            const jsonReq = await req.json();
            if (isEqual(jsonReq.filter.state, [DatumState.FAILED])) {
              return res(
                ctx.json([
                  buildDatum({
                    datum: {
                      id: '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
                    },
                    state: DatumState.FAILED,
                  }),
                ]),
              );
            }
            return res(
              ctx.json([
                buildDatum({
                  datum: {
                    id: '01db2bed340f91bc778ad9792d694f6f665e1b0dd9c7059d4f27493c1fe86155',
                  },
                  state: DatumState.FAILED,
                }),
                buildDatum({
                  datum: {
                    id: '006fdb9ba8a1afa805823336f4a280fd5c0b5c169ec48af78d07cecb96f8f14f',
                  },
                  state: DatumState.SUCCESS,
                }),
              ]),
            );
          },
        ),
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
      server.use(mockGetJob5CDatum05());
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
        rest.post<InspectJobRequest, Empty, JobInfo>(
          '/api/pps_v2.API/InspectJob',
          async (req, res, ctx) => {
            const body = await req.json();
            if (body?.job?.id === '5c1aa9bc87dd411ba5a1be0c80a3ebc2') {
              return res(
                ctx.json(
                  buildJob({
                    job: {
                      id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
                      pipeline: {name: 'montage'},
                    },
                    state: JobState.JOB_SUCCESS,
                    finished: getISOStringFromUnix(1690899618),
                    dataProcessed: '100',
                    dataSkipped: '0',
                    dataFailed: '0',
                    dataTotal: '100',
                    dataRecovered: '0',
                  }),
                ),
              );
            }
          },
        ),
      );

      const datums = generatePagingDatums({n: 100});

      server.use(
        rest.post<ListDatumRequest, Empty, DatumInfo[]>(
          '/api/pps_v2.API/ListDatum',
          async (req, res, ctx) => {
            const {paginationMarker, number} = await req.json();

            if (Number(number) === 51 && !paginationMarker) {
              return res(ctx.json(datums.slice(0, 51)));
            }

            if (
              Number(number) === 51 &&
              paginationMarker === datums[49].datum?.id
            ) {
              return res(ctx.json(datums.slice(50)));
            }

            return res(ctx.json(datums));
          },
        ),
      );

      render(<LeftPanel />);
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

      const forwards = (await screen.findAllByTestId('Pager__forward'))[0];
      expect(forwards).toBeEnabled();
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
        rest.post<InspectJobRequest, Empty, JobInfo>(
          '/api/pps_v2.API/InspectJob',
          async (req, res, ctx) => {
            const body = await req.json();
            if (body?.job?.id === '5c1aa9bc87dd411ba5a1be0c80a3ebc2') {
              return res(
                ctx.json(
                  buildJob({
                    job: {
                      id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
                      pipeline: {name: 'montage'},
                    },
                    state: JobState.JOB_SUCCESS,
                    finished: getISOStringFromUnix(1690899618),
                    dataProcessed: '2',
                    dataRecovered: '0',
                    dataSkipped: '1',
                    dataFailed: '1',
                    dataTotal: '4',
                  }),
                ),
              );
            }
          },
        ),
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
        rest.post<InspectJobRequest, Empty, JobInfo>(
          '/api/pps_v2.API/InspectJob',
          (_req, res, ctx) => {
            return res(
              ctx.json(
                buildJob({
                  job: {
                    id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
                    pipeline: {name: 'montage'},
                  },
                  state: JobState.JOB_RUNNING,
                }),
              ),
            );
          },
        ),
      );
      render(<LeftPanel />);

      expect(
        await screen.findByTestId('DatumList__processing'),
      ).toHaveTextContent('Processing — datums are being processed.');
    });
  });
});
