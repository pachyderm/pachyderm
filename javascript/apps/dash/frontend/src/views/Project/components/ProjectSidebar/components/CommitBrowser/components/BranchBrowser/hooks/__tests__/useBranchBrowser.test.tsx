import {render} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import useBranchBrowser from '../useBranchBrowser';

const BranchBrowserComponent: React.FC = withContextProviders(() => {
  const {dropdownItems} = useBranchBrowser({
    branches: [
      {id: 'master', name: 'master'},
      {id: 'develop', name: 'develop'},
      {id: 'feature', name: 'feature'},
      {id: 'alpha', name: 'alpha'},
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
      '/project/3/dag/cron/repo/cron/branch/master',
    );

    const {getByText} = render(<BranchBrowserComponent />);
    const items = getByText('master-none-alpha-develop-feature');

    expect(items).toBeInTheDocument();
  });
});
