import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockGetJobSet1D,
  mockGetEnterpriseInfoInactive,
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
    window.history.replaceState({}, '', '/');
    server.use(mockGetJobSet1D());
    server.use(mockGetEnterpriseInfoInactive());
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
    server.use(mockEmptyJobSet());

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
