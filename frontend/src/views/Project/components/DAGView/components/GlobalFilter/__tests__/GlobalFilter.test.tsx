import {mockJobSetQuery, JobState} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {mockGetJobSet_1D, JOBSET_1D} from '@dash-frontend/mocks';
import {click, type, withContextProviders} from '@dash-frontend/testHelpers';

import GlobalFilter from '..';

describe('Global ID Filter', () => {
  const server = setupServer();

  const GlobalIDFilter = withContextProviders(GlobalFilter);

  beforeAll(() => {
    server.listen();
  });

  beforeEach(() => {
    window.history.replaceState({}, '', '/');
    server.use(mockGetJobSet_1D());
  });

  afterAll(() => server.close());

  it('should reject non-ID input', async () => {
    render(<GlobalIDFilter />);

    await click(screen.getByRole('button', {name: 'Filter by Global ID'}));
    const input = await screen.findByRole('textbox', {name: 'Global ID'});

    await type(input, 'non-id input');

    await click(screen.getByRole('button', {name: 'Apply Global ID Filter'}));

    expect(
      await screen.findByText('Not a valid Global ID'),
    ).toBeInTheDocument();
  });

  it('should reject invalid jobset IDs', async () => {
    server.use(
      mockJobSetQuery((_req, res, ctx) => {
        return res(
          ctx.data({
            jobSet: {
              id: 'aaaaaaaaaaaa4aaaaaaaaaaaaaaaaaaa',
              state: JobState.JOB_SUCCESS,
              createdAt: null,
              startedAt: null,
              finishedAt: null,
              inProgress: false,
              jobs: [],
              __typename: 'JobSet',
            },
          }),
        );
      }),
    );

    render(<GlobalIDFilter />);

    await click(screen.getByRole('button', {name: 'Filter by Global ID'}));
    const input = await screen.findByRole('textbox', {name: 'Global ID'});

    await type(input, 'aaaaaaaaaaaa4aaaaaaaaaaaaaaaaaaa');
    await click(screen.getByRole('button', {name: 'Apply Global ID Filter'}));

    expect(
      await screen.findByText('This Global ID does not exist'),
    ).toBeInTheDocument();
  });

  it('should apply a valid jobset ID', async () => {
    render(<GlobalIDFilter />);

    await click(screen.getByRole('button', {name: 'Filter by Global ID'}));
    const input = await screen.findByRole('textbox', {name: 'Global ID'});

    await type(input, '1dc67e479f03498badcc6180be4ee6ce');
    await click(screen.getByRole('button', {name: 'Apply Global ID Filter'}));

    expect(
      screen.getByRole('button', {name: 'Global ID: 1dc67e...'}),
    ).toBeInTheDocument();
    expect(window.location.search).toBe(
      '?globalIdFilter=1dc67e479f03498badcc6180be4ee6ce',
    );
  });

  it('should clear an applied global filter', async () => {
    server.use(
      mockJobSetQuery((req, res, ctx) => {
        const {id} = req.variables.args;
        if (id === '1dc67e479f03498badcc6180be4ee6ce') {
          return res(
            ctx.data({
              jobSet: JOBSET_1D,
            }),
          );
        }
        return res(
          ctx.data({
            jobSet: {
              id: '',
              state: JobState.JOB_SUCCESS,
              createdAt: null,
              startedAt: null,
              finishedAt: null,
              inProgress: false,
              jobs: [],
              __typename: 'JobSet',
            },
          }),
        );
      }),
    );

    render(<GlobalIDFilter />);

    await click(screen.getByRole('button', {name: 'Filter by Global ID'}));
    const input = await screen.findByRole('textbox', {name: 'Global ID'});

    await type(input, '1dc67e479f03498badcc6180be4ee6ce');

    await click(screen.getByRole('button', {name: 'Apply Global ID Filter'}));
    await click(screen.getByRole('button', {name: 'Global ID: 1dc67e...'}));
    await click(screen.getByRole('button', {name: 'Clear Global ID filter'}));

    expect(
      screen.getByRole('button', {name: 'Filter by Global ID'}),
    ).toBeInTheDocument();
    expect(window.location.search).toBe('');
  });
});
