import {render, screen} from '@testing-library/react';
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
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/0918ac9d5daa76b86e3bb5e88e4c43a4/html_pachyderm.html',
    );
    render(<FileBrowser />);

    const iframe = await screen.findByTestId(
      'WebPreview__iframe',
      {},
      {timeout: 10000},
    );

    expect(iframe).toContainHTML('Alfreds Futterkiste');
    expect(iframe).toContainHTML('Magazzini Alimentari Riuniti');
  });
});
