import {render, screen} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import useBranchBrowser from '../useBranchBrowser';

const BranchBrowserComponent: React.FC = withContextProviders(() => {
  const {dropdownItems} = useBranchBrowser({
    branches: [
      {name: 'master'},
      {name: 'develop'},
      {name: 'feature'},
      {name: 'alpha'},
    ],
  });
  const items = dropdownItems.map((item) => item.value).join('-');

  return <span>{items}</span>;
});

describe('BranchBrowser/hooks/useBranchBrowser', () => {
  it('should sort branches with master on top', () => {
    window.history.replaceState(
      '',
      '',
      '/lineage/Solar-Power-Data-Logger-Team-Collab/repos/cron/branch/master',
    );

    render(<BranchBrowserComponent />);
    const items = screen.getByText('master-alpha-develop-feature');

    expect(items).toBeInTheDocument();
  });
});
