import {
  render,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react';
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
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/html_pachyderm.html',
    );
    render(<FileBrowser />);

    await waitForElementToBeRemoved(() =>
      screen.queryAllByLabelText('loading'),
    );

    const iframe = await screen.findByTestId(
      'WebPreview__iframe',
      {},
      {timeout: 10000},
    );

    expect(iframe).toContainHTML('Alfreds Futterkiste');
    expect(iframe).toContainHTML('Magazzini Alimentari Riuniti');
  });
});
