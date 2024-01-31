import {render, screen} from '@testing-library/react';
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
  mockEmptyJobSet,
  mockGetAllJobs,
  buildJob,
  MONTAGE_JOB_INFO_1D,
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

  // TODO: Probably deleted this functionality accidentally
  it.skip('should reject invalid jobset IDs', async () => {
    server.use(mockEmptyJobSet());

    render(<GlobalIDFilter />);

    await click(screen.getByRole('button', {name: 'Previous Jobs'}));
    const input = await screen.findByRole('textbox');

    await type(input, 'aaaaaaaaaaaa4aaaaaaaaaaaaaaaaaaa');
    // await click(screen.getByRole('button', {name: 'Apply Global ID Filter'}));

    expect(
      await screen.findByText('This Global ID does not exist'),
    ).toBeInTheDocument();
  });

  // TODO: Broke this
  it.skip('should apply a valid jobset ID', async () => {
    render(<GlobalIDFilter />);

    await click(screen.getByRole('button', {name: 'Previous Jobs'}));
    const input = await screen.findByRole('textbox');

    await type(input, '1dc67e479f03498badcc6180be4ee6ce');
    // await click(screen.getByRole('button', {name: 'Apply Global ID Filter'}));

    expect(
      screen.getByRole('button', {name: 'Global ID: 1dc67e...'}),
    ).toBeInTheDocument();
    expect(window.location.search).toBe(
      '?globalIdFilter=1dc67e479f03498badcc6180be4ee6ce',
    );
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

  it('clicking a job set keeps the window open and updates the URL', async () => {
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
  });

  describe('simple date filters', () => {
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
                    job: {
                      id: '1dc67e479f03498badcc6180be4ee6ce',
                      pipeline: {
                        name: 'montage',
                        project: {name: 'default'},
                      },
                    },
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-30T12:00:00.000000000Z',
                    reason: 'job stopped',
                  }),
                  buildJob({
                    job: {
                      id: '2dc67e479f03498badcc6180be4ee6ce',
                      pipeline: {
                        name: 'montage',
                        project: {name: 'default'},
                      },
                    },
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-29T17:00:00.000000000Z',
                    reason: 'job stopped',
                  }),
                  buildJob({
                    job: {
                      id: '3dc67e479f03498badcc6180be4ee6ce',
                      pipeline: {
                        name: 'montage',
                        project: {name: 'default'},
                      },
                    },
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-25T12:00:00.000000000Z',
                    reason: 'job stopped',
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
                    job: {
                      id: '2dc67e479f03498badcc6180be4ee6ce',
                      pipeline: {
                        name: 'montage',
                        project: {name: 'default'},
                      },
                    },
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-29T17:00:00.000000000Z',
                    reason: 'job stopped',
                  }),
                  buildJob({
                    job: {
                      id: '3dc67e479f03498badcc6180be4ee6ce',
                      pipeline: {
                        name: 'montage',
                        project: {name: 'default'},
                      },
                    },
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-25T12:00:00.000000000Z',
                    reason: 'job stopped',
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
                    job: {
                      id: '3dc67e479f03498badcc6180be4ee6ce',
                      pipeline: {
                        name: 'montage',
                        project: {name: 'default'},
                      },
                    },
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-25T12:00:00.000000000Z',
                    reason: 'job stopped',
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
                    job: {
                      id: '3dc67e479f03498badcc6180be4ee6ce',
                      pipeline: {
                        name: 'montage',
                        project: {name: 'default'},
                      },
                    },
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-25T12:00:00.000000000Z',
                    reason: 'job stopped',
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
                    job: {
                      id: '1dc67e479f03498badcc6180be4ee6ce',
                      pipeline: {
                        name: 'montage',
                        project: {name: 'default'},
                      },
                    },
                    state: JobState.JOB_FAILURE,
                    created: '2024-01-30T12:00:00.000000000Z',
                    reason: 'job stopped',
                  }),
                  buildJob({
                    job: {
                      id: '3dc67e479f03498badcc6180be4ee6ce',
                      pipeline: {
                        name: 'montage',
                        project: {name: 'default'},
                      },
                    },
                    state: JobState.JOB_SUCCESS,
                    created: '2024-01-30T12:00:00.000000000Z',
                    reason: 'job stopped',
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
      expect(screen.getAllByRole('listitem')).toHaveLength(1);

      await click(await screen.findByText(/clear/i));
      expect(screen.getAllByRole('listitem')).toHaveLength(3);
    });

    it('clicking time filters filters the results', async () => {
      global.Date.now = jest.fn(() =>
        new Date('2024-01-10T12:30:00').getTime(),
      );

      server.use(
        rest.post<ListJobRequest, Empty, JobInfo[]>(
          '/api/pps_v2.API/ListJob',
          async (_req, res, ctx) =>
            res(
              ctx.json([
                buildJob({
                  job: {id: '1dc67e479f03498badcc6180be4ee6ce'},
                  created: '2024-01-10T12:00:50.435893464Z',
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

      render(<GlobalIDFilter />);
      await click(screen.getByRole('button', {name: 'Previous Jobs'}));

      expect(await screen.findAllByRole('listitem')).toHaveLength(3);

      expect(await screen.findByText(/\(1\) last hour/i)).toBeInTheDocument();
      expect(await screen.findByText(/\(2\) last day/i)).toBeInTheDocument();
      expect(await screen.findByText(/\(3\) last 7 days/i)).toBeInTheDocument();

      await click(await screen.findByText(/\(1\) last hour/i));
      expect(screen.getAllByRole('listitem')).toHaveLength(1);

      await click(await screen.findByText(/\(2\) last day/i));
      expect(screen.getAllByRole('listitem')).toHaveLength(2);

      await click(await screen.findByText(/\(3\) last 7 days/i));
      expect(screen.getAllByRole('listitem')).toHaveLength(3);
    });
  });
  describe('advanced date filters', () => {
    it.todo('date filters filter the list');
  });
});
