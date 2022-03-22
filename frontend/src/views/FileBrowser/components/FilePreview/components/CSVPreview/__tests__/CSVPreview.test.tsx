import {render, waitFor} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';
import FileBrowserComponent from '@dash-frontend/views/FileBrowser';

describe('CSV Preview', () => {
  const FileBrowser = withContextProviders(() => {
    return <FileBrowserComponent />;
  });

  it('should default display of comma delimited files', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/3/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/csv_commas.csv',
    );
    const {findByText, queryByTestId} = render(<FileBrowser />);

    await waitFor(() =>
      expect(queryByTestId('JobDetails__loading')).not.toBeInTheDocument(),
    );
    expect(await findByText('Separator: comma')).toBeInTheDocument();
    expect(await findByText('"a"')).toBeInTheDocument();
  });

  it('should default display of tab delimited files', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/3/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/csv_tabs.csv',
    );
    const {findByText, queryByTestId} = render(<FileBrowser />);

    await waitFor(() =>
      expect(queryByTestId('JobDetails__loading')).not.toBeInTheDocument(),
    );
    expect(await findByText('Separator: tab')).toBeInTheDocument();
    expect(await findByText('"a"')).toBeInTheDocument();
  });

  it('should support .tsv file extensions', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/3/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/tsv_tabs.tsv',
    );
    const {findByText, queryByTestId} = render(<FileBrowser />);

    await waitFor(() =>
      expect(queryByTestId('JobDetails__loading')).not.toBeInTheDocument(),
    );
    expect(await findByText('Separator: tab')).toBeInTheDocument();
    expect(await findByText('"a"')).toBeInTheDocument();
  });
});
