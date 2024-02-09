import {render, screen, waitFor} from '@testing-library/react';
import {rest} from 'msw';
import {setupServer} from 'msw/node';
import React from 'react';

import {Empty} from '@dash-frontend/api/googleTypes';
import {
  InspectJobSetRequest,
  JobInfo,
  JobState,
  ListJobRequest,
} from '@dash-frontend/api/pps';
import {
  mockGetEnterpriseInfoInactive,
  mockGetAllJobs,
  buildJob,
  MONTAGE_JOB_INFO_1D,
  mockEmptyJobSet,
} from '@dash-frontend/mocks';
import {click, type, withContextProviders} from '@dash-frontend/testHelpers';

import GlobalFilter from '..';

describe('Global ID Filter', () => {
  const server = setupServer();

  const GlobalIDFilter = withContextProviders(GlobalFilter);

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    window.history.replaceState({}, '', '/lineage/default');
    server.resetHandlers();
    // server.use(mockGetJobSet1D());
    server.use(mockGetAllJobs());
    server.use(mockGetEnterpriseInfoInactive());
    server.use(
      rest.post<InspectJobSetRequest, Empty, JobInfo[]>(
        '/api/pps_v2.API/InspectJobSet',
        async (_req, res, ctx) => res(ctx.json([MONTAGE_JOB_INFO_1D])),
      ),
    );
  });
  afterAll(() => server.close());

  it('should reject non-ID input', async () => {
    render(<GlobalIDFilter />);

    await click(screen.getByRole('button', {name: 'Previous Jobs'}));
    const input = await screen.findByRole('textbox');

    await type(input, 'non-id input');

    // await click(screen.getByRole('button', {name: 'Apply Global ID Filter'}));

    expect(
      await screen.findByText('Not a valid Global ID'),
    ).toBeInTheDocument();
  });

  it('should reject invalid jobset IDs', async () => {
    server.use(mockEmptyJobSet());

    render(<GlobalIDFilter />);

    await click(screen.getByRole('button', {name: 'Previous Jobs'}));
    const input = await screen.findByRole('textbox');

    await type(input, 'aaaaaaaaaaaa4aaaaaaaaaaaaaaaaaaa');
    await click(await screen.findByRole('button', {name: /Apply Filter/i}));

    expect(
      await screen.findByText('This Global ID does not exist'),
    ).toBeInTheDocument();
  });

  it('should reject jobset IDs from different projects', async () => {
    server.use(
      rest.post<InspectJobSetRequest, Empty, JobInfo[]>(
        '/api/pps_v2.API/InspectJobSet',
        async (req, res, ctx) => {
          const body = await req.json();
          if (body?.jobSet?.id === '5c1aa9bc87dd411ba5a1be0c80a3ebc2') {
            return res(
              ctx.json([
                buildJob({
                  job: {
                    id: '5c1aa9bc87dd411ba5a1be0c80a3ebc2',
                    pipeline: {
                      name: 'montage',
                      project: {name: 'test'},
                    },
                  },
                }),
              ]),
            );
          }
        },
      ),
    );

    render(<GlobalIDFilter />);

    await click(screen.getByRole('button', {name: 'Previous Jobs'}));
    const input = await screen.findByRole('textbox');

    await type(input, '5c1aa9bc87dd411ba5a1be0c80a3ebc2');
    await click(await screen.findByRole('button', {name: /Apply Filter/i}));

    expect(
      await screen.findByText('This Global ID from a different project'),
    ).toBeInTheDocument();
  });

  it('should error when given an invalid global ID', async () => {
    render(<GlobalIDFilter />);

    await click(screen.getByRole('button', {name: 'Previous Jobs'}));
    const input = await screen.findByRole('textbox');

    await type(input, '11');

    expect(
      await screen.findByText('Not a valid Global ID'),
    ).toBeInTheDocument();
  });

  it('should apply a valid global id then the button should change to clear filter', async () => {
    render(<GlobalIDFilter />);

    await click(screen.getByRole('button', {name: 'Previous Jobs'}));
    const input = await screen.findByRole('textbox');

    await type(input, '1dc67e479f03498badcc6180be4ee6ce');
    await click(await screen.findByRole('button', {name: /Apply Filter/i}));

    expect(
      await screen.findByRole('button', {name: /viewing/i}),
    ).toBeInTheDocument();
    expect(window.location.search).toBe(
      '?globalIdFilter=1dc67e479f03498badcc6180be4ee6ce',
    );

    await click(await screen.findByRole('button', {name: /Clear Filter/i}));
  });

  it('should close side panel when applying a global ID', async () => {
    window.history.replaceState(
      {},
      '',
      '/lineage/default/pipelines/dreamy-gnu',
    );

    render(<GlobalIDFilter />);
    expect(window.location.pathname).toBe(
      '/lineage/default/pipelines/dreamy-gnu',
    );

    await click(screen.getByRole('button', {name: 'Previous Jobs'}));
    const input = await screen.findByRole('textbox');

    await type(input, '1dc67e479f03498badcc6180be4ee6ce');
    await click(await screen.findByRole('button', {name: /Apply Filter/i}));

    expect(
      await screen.findByRole('button', {name: /viewing/i}),
    ).toBeInTheDocument();
    expect(window.location.pathname).toBe('/lineage/default');
  });

  it('should display all job sets', async () => {
    render(<GlobalIDFilter />);
    await click(screen.getByRole('button', {name: 'Previous Jobs'}));

    expect(await screen.findByRole('list')).toBeInTheDocument();
    expect(await screen.findAllByRole('listitem')).toHaveLength(5);
  });

  it('should display passing and failing state of job sets', async () => {
    render(<GlobalIDFilter />);
    await click(screen.getByRole('button', {name: 'Previous Jobs'}));

    expect(
      await screen.findAllByTestId('GlobalFilter__status_running'),
    ).toHaveLength(1);
    expect(
      await screen.findAllByTestId('GlobalFilter__status_passing'),
    ).toHaveLength(3);
    expect(
      await screen.findAllByTestId('GlobalFilter__status_failing'),
    ).toHaveLength(1);
  });

  it('clicking a job set keeps the window open, updates the URL, and updates the text in the button', async () => {
    render(<GlobalIDFilter />);
    await click(screen.getByRole('button', {name: 'Previous Jobs'}));

    await click(await screen.findByText(/1dc67e479f03498badcc6180be4ee6ce/));
    expect(window.location.search).toBe(
      '?globalIdFilter=1dc67e479f03498badcc6180be4ee6ce',
    );

    await click(await screen.findByText(/a4423427351e42aabc40c1031928628e/));
    expect(window.location.search).toBe(
      '?globalIdFilter=a4423427351e42aabc40c1031928628e',
    );

    expect(
      await screen.findByRole('button', {name: /Viewing Aug 1, 2023; 14:20/i}),
    ).toBeInTheDocument();
  });

  describe('Simple date filter', () => {
    afterEach(() => {
      jest.restoreAllMocks();
    });

    describe('should display the correct default date filter buttons', () => {
      it('should have the correct default date filter buttons - all', async () => {
        server.use(
          rest.post<ListJobRequest, Empty, JobInfo[]>(
            '/api/pps_v2.API/ListJob',
            async (_req, res, ctx) =>
              res(
                ctx.json([
                  buildJob({
                    job: {id: '1dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-30T12:00:00.000000000Z',
                  }),
                  buildJob({
                    job: {id: '2dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-29T17:00:00.000000000Z',
                  }),
                  buildJob({
                    job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-25T12:00:00.000000000Z',
                  }),
                ]),
              ),
          ),
        );

        global.Date.now = jest.fn(() =>
          new Date('2024-01-30T12:01:00').getTime(),
        );

        render(<GlobalIDFilter />);
        await click(screen.getByRole('button', {name: 'Previous Jobs'}));

        expect(await screen.findByText(/\(1\) last hour/i)).toBeInTheDocument();
        expect(await screen.findByText(/\(2\) last day/i)).toBeInTheDocument();
        expect(
          await screen.findByText(/\(3\) last 7 days/i),
        ).toBeInTheDocument();
      });

      it('should have the correct default date filter buttons - no hour', async () => {
        server.use(
          rest.post<ListJobRequest, Empty, JobInfo[]>(
            '/api/pps_v2.API/ListJob',
            async (req, res, ctx) =>
              res(
                ctx.json([
                  buildJob({
                    job: {id: '2dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-29T17:00:00.000000000Z',
                  }),
                  buildJob({
                    job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-25T12:00:00.000000000Z',
                  }),
                ]),
              ),
          ),
        );

        global.Date.now = jest.fn(() =>
          new Date('2024-01-30T10:00:00').getTime(),
        );

        render(<GlobalIDFilter />);
        await click(screen.getByRole('button', {name: 'Previous Jobs'}));

        expect(await screen.findByText(/\(1\) last day/i)).toBeInTheDocument();
        expect(
          await screen.findByText(/\(2\) last 7 days/i),
        ).toBeInTheDocument();
        expect(screen.queryByText(/last hour/i)).not.toBeInTheDocument();
      });

      it('should have the correct default date filter buttons - no hour or day', async () => {
        server.use(
          rest.post<ListJobRequest, Empty, JobInfo[]>(
            '/api/pps_v2.API/ListJob',
            async (req, res, ctx) =>
              res(
                ctx.json([
                  buildJob({
                    job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-25T12:00:00.000000000Z',
                  }),
                ]),
              ),
          ),
        );

        global.Date.now = jest.fn(() =>
          new Date('2024-01-30T10:00:00').getTime(),
        );

        render(<GlobalIDFilter />);
        await click(screen.getByRole('button', {name: 'Previous Jobs'}));

        expect(
          await screen.findByText(/\(1\) last 7 days/i),
        ).toBeInTheDocument();
        expect(screen.queryByText(/last hour/i)).not.toBeInTheDocument();
        expect(screen.queryByText(/last day/i)).not.toBeInTheDocument();
      });

      it('should have the correct default date filter buttons - none', async () => {
        server.use(
          rest.post<ListJobRequest, Empty, JobInfo[]>(
            '/api/pps_v2.API/ListJob',
            async (req, res, ctx) =>
              res(
                ctx.json([
                  buildJob({
                    job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-25T12:00:00.000000000Z',
                  }),
                ]),
              ),
          ),
        );
        global.Date.now = jest.fn(() =>
          new Date('2024-09-03T14:21:00').getTime(),
        );

        render(<GlobalIDFilter />);
        await click(screen.getByRole('button', {name: 'Previous Jobs'}));

        expect(
          await screen.findByText(/3dc67e479f03498badcc6180be4ee6ce/),
        ).toBeInTheDocument();
        expect(screen.queryByText(/last hour/i)).not.toBeInTheDocument();
        expect(screen.queryByText(/last day/i)).not.toBeInTheDocument();
        expect(
          screen.queryByText(/\(3\) last 7 days/i),
        ).not.toBeInTheDocument();
      });

      it('should not account for passing jobs', async () => {
        server.use(
          rest.post<ListJobRequest, Empty, JobInfo[]>(
            '/api/pps_v2.API/ListJob',
            async (req, res, ctx) =>
              res(
                ctx.json([
                  buildJob({
                    job: {id: '1dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-30T12:00:00.000000000Z',
                  }),
                  buildJob({
                    job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_SUCCESS,
                    created: '2024-01-30T12:00:00.000000000Z',
                  }),
                ]),
              ),
          ),
        );

        global.Date.now = jest.fn(() =>
          new Date('2024-01-30T10:00:00').getTime(),
        );

        render(<GlobalIDFilter />);
        await click(screen.getByRole('button', {name: 'Previous Jobs'}));

        expect(await screen.findByText(/\(1\) last hour/i)).toBeInTheDocument();
      });
    });

    it('clear button works', async () => {
      server.use(
        rest.post<ListJobRequest, Empty, JobInfo[]>(
          '/api/pps_v2.API/ListJob',
          async (_req, res, ctx) =>
            res(
              ctx.json([
                buildJob({
                  job: {id: '1dc67e479f03498badcc6180be4ee6ce'},
                  created: '2024-01-25T12:00:00.000000000Z',
                  state: JobState.JOB_FAILURE,
                }),
                buildJob({
                  job: {id: '2dc67e479f03498badcc6180be4ee6ce'},
                  created: '2024-01-10T08:31:50.435893464Z',
                  state: JobState.JOB_FAILURE,
                }),
                buildJob({
                  job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                  created: '2024-01-08T18:31:50.435893464Z',
                  state: JobState.JOB_FAILURE,
                }),
              ]),
            ),
        ),
      );

      global.Date.now = jest.fn(() =>
        new Date('2024-01-10T12:30:00').getTime(),
      );

      render(<GlobalIDFilter />);
      await click(screen.getByRole('button', {name: 'Previous Jobs'}));

      expect(await screen.findAllByRole('listitem')).toHaveLength(3);

      expect(await screen.findByText(/\(1\) last hour/i)).toBeInTheDocument();
      expect(await screen.findByText(/\(2\) last day/i)).toBeInTheDocument();
      expect(await screen.findByText(/\(3\) last 7 days/i)).toBeInTheDocument();

      await click(await screen.findByText(/\(1\) last hour/i));
      await waitFor(() => {
        expect(screen.getAllByRole('listitem')).toHaveLength(1);
      });

      await click(await screen.findByText(/clear/i));
      await waitFor(() => {
        expect(screen.getAllByRole('listitem')).toHaveLength(3);
      });
    });

    it('clicking time filters filters the results', async () => {
      server.use(
        rest.post<ListJobRequest, Empty, JobInfo[]>(
          '/api/pps_v2.API/ListJob',
          async (_req, res, ctx) =>
            res(
              ctx.json([
                buildJob({
                  job: {id: '1dc67e479f03498badcc6180be4ee6ce'},
                  created: '2024-01-25T12:00:00.000000000Z',
                  state: JobState.JOB_FAILURE,
                }),
                buildJob({
                  job: {id: '2dc67e479f03498badcc6180be4ee6ce'},
                  created: '2024-01-10T08:31:50.435893464Z',
                  state: JobState.JOB_FAILURE,
                }),
                buildJob({
                  job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                  created: '2024-01-08T18:31:50.435893464Z',
                  state: JobState.JOB_FAILURE,
                }),
              ]),
            ),
        ),
      );

      global.Date.now = jest.fn(() =>
        new Date('2024-01-10T12:30:00').getTime(),
      );

      render(<GlobalIDFilter />);
      await click(screen.getByRole('button', {name: 'Previous Jobs'}));

      expect(await screen.findAllByRole('listitem')).toHaveLength(3);

      expect(await screen.findByText(/\(1\) last hour/i)).toBeInTheDocument();
      expect(await screen.findByText(/\(2\) last day/i)).toBeInTheDocument();
      expect(await screen.findByText(/\(3\) last 7 days/i)).toBeInTheDocument();

      await click(await screen.findByText(/\(1\) last hour/i));
      await waitFor(() => {
        expect(screen.getAllByRole('listitem')).toHaveLength(1);
      });

      await click(await screen.findByText(/\(2\) last day/i));
      await waitFor(() => {
        expect(screen.getAllByRole('listitem')).toHaveLength(2);
      });

      await click(await screen.findByText(/\(3\) last 7 days/i));
      await waitFor(() => {
        expect(screen.getAllByRole('listitem')).toHaveLength(3);
      });
    });
  });

  describe('Advanced date filter', () => {
    it('setting the date and time shows a readable date time string', async () => {
      render(<GlobalIDFilter />);

      await click(screen.getByRole('button', {name: 'Previous Jobs'}));
      await click(screen.getByRole('button', {name: /filter/i}));

      await type(
        screen.getByTestId('AdvancedDateTimeFilter__startDate'),
        '2024-01-10',
      );
      await type(
        screen.getByTestId('AdvancedDateTimeFilter__endDate'),
        '2024-01-10',
      );

      expect(
        await screen.findByText(
          /January 10, 2024 at 00:00 - January 10, 2024 at 23:59/i,
        ),
      ).toBeInTheDocument();

      await type(
        screen.getByTestId('AdvancedDateTimeFilter__startTime'),
        '09:30',
      );
      await type(
        screen.getByTestId('AdvancedDateTimeFilter__endTime'),
        '11:30',
      );

      expect(
        await screen.findByText(
          /January 10, 2024 at 09:30 - January 10, 2024 at 11:30/i,
        ),
      ).toBeInTheDocument();
    });

    it('shows both passing and failing jobs', async () => {
      server.use(
        rest.post<ListJobRequest, Empty, JobInfo[]>(
          '/api/pps_v2.API/ListJob',
          async (_req, res, ctx) =>
            res(
              ctx.json([
                buildJob({
                  job: {id: '2dc67e479f03498badcc6180be4ee6ce'},
                  state: JobState.JOB_SUCCESS,
                  created: '2024-01-10T17:00:00.000000000Z',
                }),
                buildJob({
                  job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                  state: JobState.JOB_FAILURE,
                  created: '2024-01-10T12:00:00.000000000Z',
                }),
              ]),
            ),
        ),
      );

      global.Date.now = jest.fn(() =>
        new Date('2024-01-10T12:01:00').getTime(),
      );

      render(<GlobalIDFilter />);

      await click(screen.getByRole('button', {name: 'Previous Jobs'}));
      await click(screen.getByRole('button', {name: /filter/i}));

      await type(
        screen.getByTestId('AdvancedDateTimeFilter__startDate'),
        '2024-01-10',
      );
      await type(
        screen.getByTestId('AdvancedDateTimeFilter__endDate'),
        '2024-01-10',
      );

      expect(await screen.findAllByRole('listitem')).toHaveLength(2);
      expect(
        await screen.findAllByTestId('GlobalFilter__status_passing'),
      ).toHaveLength(1);
      expect(
        await screen.findAllByTestId('GlobalFilter__status_failing'),
      ).toHaveLength(1);
    });

    it('setting a start and end date filter shows the correct items', async () => {
      server.use(
        rest.post<ListJobRequest, Empty, JobInfo[]>(
          '/api/pps_v2.API/ListJob',
          async (req, res, ctx) => {
            const jsonReq = await req.json();
            if (
              jsonReq.jqFilter ===
              'select((.created | (split(".")[0] + "Z") | fromdateiso8601 >= 1704844800))'
            ) {
              return res(
                ctx.json([
                  buildJob({
                    job: {id: '1dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                  }),
                ]),
              );
            } else if (
              jsonReq.jqFilter ===
              'select((.created | (split(".")[0] + "Z") | fromdateiso8601 >= 1704844800) and (.created | (split(".")[0] + "Z") | fromdateiso8601 <= 1704931199))'
            ) {
              return res(
                ctx.json([
                  buildJob({
                    job: {id: '2dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                  }),
                  buildJob({
                    job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                  }),
                ]),
              );
            } else {
              return res(
                ctx.json([
                  buildJob({
                    job: {id: '1dc67e479f03498badcc6180be4ee6ce'},
                  }),
                  buildJob({
                    job: {id: '2dc67e479f03498badcc6180be4ee6ce'},
                  }),
                  buildJob({
                    job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                  }),
                ]),
              );
            }
          },
        ),
      );

      global.Date.now = jest.fn(() =>
        new Date('2024-01-10T12:01:00').getTime(),
      );

      render(<GlobalIDFilter />);

      await click(screen.getByRole('button', {name: 'Previous Jobs'}));
      await click(screen.getByRole('button', {name: /filter/i}));

      expect(await screen.findAllByRole('listitem')).toHaveLength(3);

      await type(
        screen.getByTestId('AdvancedDateTimeFilter__startDate'),
        '2024-01-10',
      );
      await waitFor(async () => {
        expect(await screen.findAllByRole('listitem')).toHaveLength(1);
      });
      expect(
        screen.getByText('1dc67e479f03498badcc6180be4ee6ce'),
      ).toBeInTheDocument();

      await type(
        screen.getByTestId('AdvancedDateTimeFilter__endDate'),
        '2024-01-10',
      );

      await waitFor(async () => {
        expect(await screen.findAllByRole('listitem')).toHaveLength(2);
      });
      expect(
        screen.getByText('2dc67e479f03498badcc6180be4ee6ce'),
      ).toBeInTheDocument();
      expect(
        screen.getByText('3dc67e479f03498badcc6180be4ee6ce'),
      ).toBeInTheDocument();
    });

    it('setting a start and end time filter shows the correct items', async () => {
      server.use(
        rest.post<ListJobRequest, Empty, JobInfo[]>(
          '/api/pps_v2.API/ListJob',
          async (req, res, ctx) => {
            const jsonReq = await req.json();
            if (
              jsonReq.jqFilter ===
              'select((.created | (split(".")[0] + "Z") | fromdateiso8601 >= 1704877380) and (.created | (split(".")[0] + "Z") | fromdateiso8601 <= 1704931199))'
            ) {
              return res(
                ctx.json([
                  buildJob({
                    job: {id: '1dc67e479f03498badcc6180be4ee6ce'},
                  }),
                  buildJob({
                    job: {id: '2dc67e479f03498badcc6180be4ee6ce'},
                  }),
                  buildJob({
                    job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                  }),
                ]),
              );
            } else if (
              jsonReq.jqFilter ===
              'select((.created | (split(".")[0] + "Z") | fromdateiso8601 >= 1704879000) and (.created | (split(".")[0] + "Z") | fromdateiso8601 <= 1704931199))'
            ) {
              return res(
                ctx.json([
                  buildJob({
                    job: {id: '2dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                  }),
                  buildJob({
                    job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                  }),
                ]),
              );
            } else if (
              jsonReq.jqFilter ===
              'select((.created | (split(".")[0] + "Z") | fromdateiso8601 >= 1704879000) and (.created | (split(".")[0] + "Z") | fromdateiso8601 <= 1704886259))'
            ) {
              return res(
                ctx.json([
                  buildJob({
                    job: {id: '2dc67e479f03498badcc6180be4ee6ce'},
                    state: JobState.JOB_FAILURE,
                  }),
                ]),
              );
            } else {
              return res(
                ctx.json([
                  buildJob({
                    job: {id: '1dc67e479f03498badcc6180be4ee6ce'},
                  }),
                  buildJob({
                    job: {id: '2dc67e479f03498badcc6180be4ee6ce'},
                  }),
                  buildJob({
                    job: {id: '3dc67e479f03498badcc6180be4ee6ce'},
                  }),
                ]),
              );
            }
          },
        ),
      );

      global.Date.now = jest.fn(() =>
        new Date('2024-01-10T12:01:00').getTime(),
      );

      render(<GlobalIDFilter />);

      await click(screen.getByRole('button', {name: 'Previous Jobs'}));
      await click(screen.getByRole('button', {name: /filter/i}));

      expect(await screen.findAllByRole('listitem')).toHaveLength(3);

      await type(
        screen.getByTestId('AdvancedDateTimeFilter__startDate'),
        '2024-01-10',
      );
      await type(
        screen.getByTestId('AdvancedDateTimeFilter__endDate'),
        '2024-01-10',
      );

      await waitFor(async () => {
        expect(await screen.findAllByRole('listitem')).toHaveLength(3);
      });

      await type(
        screen.getByTestId('AdvancedDateTimeFilter__startTime'),
        '09:30',
      );

      await waitFor(async () => {
        expect(await screen.findAllByRole('listitem')).toHaveLength(2);
      });

      await type(
        screen.getByTestId('AdvancedDateTimeFilter__endTime'),
        '11:30',
      );

      await waitFor(async () => {
        expect(await screen.findAllByRole('listitem')).toHaveLength(1);
      });
    });
  });
});
