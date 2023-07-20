import enterpriseStates from '@dash-backend/mock/fixtures/enterprise';
import {render, screen} from '@testing-library/react';
import React from 'react';

import {withContextProviders, mockServer} from '@dash-frontend/testHelpers';

import LandingHeader from '../../LandingHeader';

describe('LandingHeader', () => {
  const Header = withContextProviders(LandingHeader);

  it('should show the hpe branding when enterprise is active', async () => {
    mockServer.getState().enterprise = enterpriseStates.active;
    render(<Header />);

    await screen.findByRole('heading', {name: 'HPE ML Data Management'});
  });

  it('should show the console branding when enterprise is inactive', async () => {
    mockServer.getState().enterprise = enterpriseStates.inactive;
    render(<Header />);

    await screen.findByRole('heading', {name: 'Console'});
  });
});
