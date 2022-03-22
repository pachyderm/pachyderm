import {render, waitFor} from '@testing-library/react';
import React from 'react';

import {type, withContextProviders} from '@dash-frontend/testHelpers';

import GlobalFilter from '../';

describe('Global ID Filter', () => {
  const GlobalIDFilter = withContextProviders(() => {
    return <GlobalFilter />;
  });

  beforeEach(() => {
    window.history.replaceState({}, '', '/');
  });

  it('should reject non-ID input', async () => {
    const {getByText, findByTestId, queryByText} = render(<GlobalIDFilter />);

    getByText('Filter by Global ID').click();

    const input = await findByTestId('GlobalFilter__name');

    await type(input, 'non-id input');

    getByText('Apply').click();

    await waitFor(() =>
      expect(queryByText('Not a valid Global ID')).toBeInTheDocument(),
    );
  });

  it('should reject invalid jobset IDs', async () => {
    const {getByText, findByTestId, queryByText} = render(<GlobalIDFilter />);

    getByText('Filter by Global ID').click();

    const input = await findByTestId('GlobalFilter__name');

    await type(input, 'aaaaaaaaaaaa4aaaaaaaaaaaaaaaaaaa');

    getByText('Apply').click();

    await waitFor(() =>
      expect(queryByText('This Global ID does not exist')).toBeInTheDocument(),
    );
  });

  it('should apply a valid jobset ID', async () => {
    const {getByText, findByTestId, queryByText} = render(<GlobalIDFilter />);

    getByText('Filter by Global ID').click();

    const input = await findByTestId('GlobalFilter__name');

    await type(input, '23b9af7d5d4343219bc8e02ff44cd55a');

    getByText('Apply').click();

    await waitFor(() =>
      expect(queryByText('Global ID: 23b9af7d')).toBeInTheDocument(),
    );
    expect(window.location.search).toEqual(
      '?view=eyJnbG9iYWxJZEZpbHRlciI6IjIzYjlhZjdkNWQ0MzQzMjE5YmM4ZTAyZmY0NGNkNTVhIn0%3D',
    );
  });

  it('should clear an applied global filter', async () => {
    const {getByText, findByTestId, queryByText} = render(<GlobalIDFilter />);

    getByText('Filter by Global ID').click();

    const input = await findByTestId('GlobalFilter__name');

    await type(input, '23b9af7d5d4343219bc8e02ff44cd55a');

    getByText('Apply').click();

    await waitFor(() =>
      expect(queryByText('Global ID: 23b9af7d')).toBeInTheDocument(),
    );

    getByText('Global ID: 23b9af7d').click();

    getByText('Clear ID').click();

    await waitFor(() =>
      expect(queryByText('Filter by Global ID')).toBeInTheDocument(),
    );
    expect(window.location.search).toEqual('?view=e30%3D');
  });
});
