import {
  render,
  screen,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';
import FileBrowserComponent from '@dash-frontend/views/FileBrowser';

describe('Markdown Preview', () => {
  const FileBrowser = withContextProviders(() => {
    return <FileBrowserComponent />;
  });

  it('should support markdown files', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/markdown_basic.md',
    );
    render(<FileBrowser />);

    await waitForElementToBeRemoved(() => screen.queryAllByRole('status'));

    // Should render the markdown contents
    expect(await screen.findByText('Installation')).toBeInTheDocument();
    expect(await screen.findByText('make install')).toBeInTheDocument();

    // Should not render the source
    expect(screen.queryByText('###')).not.toBeInTheDocument();
  });
});
