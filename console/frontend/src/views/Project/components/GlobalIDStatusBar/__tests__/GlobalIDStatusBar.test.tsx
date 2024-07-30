import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockGetEnterpriseInfoInactive,
  mockGetJobSets,
} from '@dash-frontend/mocks';
import {click, withContextProviders} from '@dash-frontend/testHelpers';

import GlobalIDStatusBarComponent from '../GlobalIDStatusBar';

describe('Global ID status bar', () => {
  const server = setupServer();

  const GlobalIDStatusBar = withContextProviders(() => {
    return <GlobalIDStatusBarComponent />;
  });

  beforeAll(() => {
    server.listen();
    server.use(mockGetJobSets());

    server.use(mockGetEnterpriseInfoInactive());
  });

  beforeEach(() => {
    window.history.replaceState(
      '',
      '',
      '/lineage/default/pipelines/montage?globalIdFilter=a4423427351e42aabc40c1031928628e',
    );
  });

  afterAll(() => server.close());

  it('should show if globalIdFilter is in the url', async () => {
    render(<GlobalIDStatusBar />);
    expect(await screen.findByText(/global id/i)).toBeInTheDocument();
    expect(await screen.findByText('a44234...')).toBeInTheDocument();
  });

  it('should close on button click and should not show if globalIdFilter is not in the url', async () => {
    render(<GlobalIDStatusBar />);
    expect(await screen.findByText('a44234...')).toBeInTheDocument();

    expect(window.location.search).toBe(
      '?globalIdFilter=a4423427351e42aabc40c1031928628e',
    );
    await click(screen.getByText('Clear Filter'));
    expect(window.location.search).toBe('');
    expect(screen.queryByText('a44234...')).not.toBeInTheDocument();
  });

  it('should display subjob states and route to the subjobs table', async () => {
    render(<GlobalIDStatusBar />);

    await click(await screen.findByText('4 s'));

    expect(window.location.pathname).toBe('/project/default/jobs/runtimes');
    expect(window.location.search).toContain(
      '?selectedJobs=a4423427351e42aabc40c1031928628e',
    );
  });

  it('should route to the runtime chart for the selected global ID', async () => {
    render(<GlobalIDStatusBar />);

    await click(await screen.findByText('2 Subjobs Total'));

    expect(window.location.pathname).toBe('/project/default/jobs/subjobs');
    expect(window.location.search).toContain(
      '?selectedJobs=a4423427351e42aabc40c1031928628e',
    );
  });

  it('should show error state if selected global ID has an error', async () => {
    render(<GlobalIDStatusBar />);
    expect(await screen.findByText('a44234...')).toBeInTheDocument();
    expect(
      await screen.findByLabelText('Global ID contains subjobs with errors'),
    ).toBeInTheDocument();
  });

  it('should not show error state if selected global ID has no errors', async () => {
    window.history.replaceState(
      '',
      '',
      '/lineage/default/pipelines/montage?globalIdFilter=1dc67e479f03498badcc6180be4ee6ce',
    );
    render(<GlobalIDStatusBar />);
    expect(await screen.findByText('1dc67e...')).toBeInTheDocument();
    expect(
      screen.queryByLabelText('Global ID contains subjobs with errors'),
    ).not.toBeInTheDocument();
  });
});
