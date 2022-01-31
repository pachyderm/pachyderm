import {render} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';
import FileBrowserComponent from '@dash-frontend/views/FileBrowser';

describe('Web Preview', () => {
  const FileBrowser = withContextProviders(() => {
    return <FileBrowserComponent />;
  });

  it('should support .html file extensions', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/3/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/html_pachyderm.html',
    );
    const {findByTestId} = render(<FileBrowser />);

    const iframe = await findByTestId('WebPreview__iframe');

    expect(iframe).toContainHTML('Alfreds Futterkiste');
    expect(iframe).toContainHTML('Magazzini Alimentari Riuniti');
  });
});
