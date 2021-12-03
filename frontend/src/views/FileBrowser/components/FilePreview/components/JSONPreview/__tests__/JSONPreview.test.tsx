import {render} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';
import FileBrowserComponent from '@dash-frontend/views/FileBrowser';

describe('JSON Preview', () => {
  const FileBrowser = withContextProviders(() => {
    return <FileBrowserComponent />;
  });

  it('should support json file extensions', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/3/repo/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/json_mixed.json',
    );
    const {findByText} = render(<FileBrowser />);

    expect(await findByText('07-06-19')).toBeInTheDocument();
    expect(await findByText('231221B')).toBeInTheDocument();
  });
});
