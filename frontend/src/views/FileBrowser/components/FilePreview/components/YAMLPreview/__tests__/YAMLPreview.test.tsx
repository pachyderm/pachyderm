import {render, waitFor, screen} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';
import FileBrowserComponent from '@dash-frontend/views/FileBrowser';

describe('YAML Preview', () => {
  const FileBrowser = withContextProviders(() => {
    return <FileBrowserComponent />;
  });

  it('should support yaml file extensions', async () => {
    window.history.replaceState(
      {},
      '',
      '/project/3/repos/cron/branch/master/commit/9d5daa0918ac4c43a476b86e3bb5e88e/yml_spec.yml',
    );
    render(<FileBrowser />);

    await waitFor(() =>
      expect(
        screen.queryByTestId('YAMLPreview__loading'),
      ).not.toBeInTheDocument(),
    );
    expect(await screen.findByText(`"sentiment_words"`)).toBeInTheDocument();
    expect(
      await screen.findByText(`"elephantjones/market_sentiment:dev0.25"`),
    ).toBeInTheDocument();
  });
});
