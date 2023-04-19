import enterpriseStates from '@dash-backend/mock/fixtures/enterprise';
import {render, screen} from '@testing-library/react';
import React from 'react';

import {
  withContextProviders,
  mockServer,
  setIdTokenForAccount,
} from '@dash-frontend/testHelpers';

import LandingHeader from '../../LandingHeader';

describe('LandingHeader', () => {
  const Header = withContextProviders(LandingHeader);

  it('should show the hpe branding when enterprise is active', async () => {
    mockServer.getState().enterprise = enterpriseStates.active;
    render(<Header />);

    await screen.findByRole('heading', {name: 'HPE MLDM'});
  });

  it('should show the console branding when enterprise is inactive', async () => {
    mockServer.getState().enterprise = enterpriseStates.inactive;
    render(<Header />);

    await screen.findByRole('heading', {name: 'Console'});
  });

  it("should display the user's name", async () => {
    render(<Header />);

    expect(
      await screen.findAllByText(`Hello, ${mockServer.getAccount().name}!`),
    ).toHaveLength(2);
  });

  it("should display user's email as a fallback", async () => {
    setIdTokenForAccount({id: 'ff7', email: 'barret.wallace@avalanche.net'});

    render(<Header />);

    expect(
      await screen.findAllByText(`Hello, barret.wallace@avalanche.net!`),
    ).toHaveLength(2);
  });
});
