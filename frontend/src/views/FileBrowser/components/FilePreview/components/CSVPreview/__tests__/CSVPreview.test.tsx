import {render} from '@testing-library/react';
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
      '/project/3/repo/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/commas.csv',
    );
    const {findByText} = render(<FileBrowser />);
    expect(await findByText('Separator: comma')).toBeInTheDocument();
    expect(await findByText('"a"')).toBeInTheDocument();
  });
  it('should default display of tab delimited files', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/3/repo/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/tabs.csv',
    );
    const {findByText} = render(<FileBrowser />);
    expect(await findByText('Separator: tab')).toBeInTheDocument();
    expect(await findByText('"a"')).toBeInTheDocument();
  });
  it('should support .tsv file extensions', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/3/repo/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/tabs.tsv',
    );
    const {findByText} = render(<FileBrowser />);
    expect(await findByText('Separator: tab')).toBeInTheDocument();
    expect(await findByText('"a"')).toBeInTheDocument();
  });
});
