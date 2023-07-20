import {render, screen} from '@testing-library/react';
import React from 'react';

import {click, type, withContextProviders} from '@dash-frontend/testHelpers';

import GlobalFilter from '../';

describe('Global ID Filter', () => {
  const GlobalIDFilter = withContextProviders(() => {
    return <GlobalFilter />;
  });

  beforeEach(() => {
    window.history.replaceState({}, '', '/');
  });

  it('should reject non-ID input', async () => {
    render(<GlobalIDFilter />);

    await click(screen.getByText('Filter by Global ID'));

    const input = await screen.findByTestId('GlobalFilter__name');

    await type(input, 'non-id input');

    await click(screen.getByText('Apply'));

    expect(
      await screen.findByText('Not a valid Global ID'),
    ).toBeInTheDocument();
  });

  it('should reject invalid jobset IDs', async () => {
    render(<GlobalIDFilter />);

    screen.getByText('Filter by Global ID').click();

    const input = await screen.findByTestId('GlobalFilter__name');

    await type(input, 'aaaaaaaaaaaa4aaaaaaaaaaaaaaaaaaa');

    screen.getByText('Apply').click();

    expect(
      await screen.findByText('This Global ID does not exist'),
    ).toBeInTheDocument();
  });

  it('should apply a valid jobset ID', async () => {
    render(<GlobalIDFilter />);

    screen.getByText('Filter by Global ID').click();

    const input = await screen.findByTestId('GlobalFilter__name');

    await type(input, '23b9af7d5d4343219bc8e02ff44cd55a');

    screen.getByText('Apply').click();

    await screen.findByText('23b9af...');
    expect(window.location.search).toBe(
      '?globalIdFilter=23b9af7d5d4343219bc8e02ff44cd55a',
    );
  });

  it('should clear an applied global filter', async () => {
    render(<GlobalIDFilter />);

    screen.getByText('Filter by Global ID').click();

    const input = await screen.findByTestId('GlobalFilter__name');

    await type(input, '23b9af7d5d4343219bc8e02ff44cd55a');

    screen.getByText('Apply').click();

    await screen.findByText('23b9af...');

    screen.getByText('23b9af...').click();

    screen.getByText('Clear ID').click();

    await screen.findByText('Filter by Global ID');
    expect(window.location.search).toBe('');
  });
});
