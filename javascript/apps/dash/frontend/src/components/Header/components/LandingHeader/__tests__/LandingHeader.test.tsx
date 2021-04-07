import {render} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';

import LandingHeaderComponent from '../LandingHeader';

describe('LandingHeader', () => {
  const LandingHeader = withContextProviders(LandingHeaderComponent);

  it('should grab workspace name from local storage', async () => {
    window.localStorage.setItem('workspaceName', 'Elegant Elephant');

    const {queryByText} = render(<LandingHeader />);

    expect(queryByText('Workspace Elegant Elephant')).toBeInTheDocument();
  });

  it('should display message when workspace not found', async () => {
    window.localStorage.removeItem('workspaceName');

    const {queryByText} = render(<LandingHeader />);

    expect(queryByText('Workspace')).not.toBeInTheDocument();
  });
});
