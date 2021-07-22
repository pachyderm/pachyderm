import {render} from '@testing-library/react';
import React from 'react';

import {
  withContextProviders,
  mockServer,
  setIdTokenForAccount,
} from '@dash-frontend/testHelpers';

import LandingHeader from '../../LandingHeader';

describe('LandingHeader', () => {
  const Header = withContextProviders(LandingHeader);

  it('should grab workspace name from local storage', async () => {
    window.localStorage.setItem('workspaceName', 'Elegant Elephant');

    const {queryByText} = render(<Header />);

    expect(queryByText('Workspace Elegant Elephant')).toBeInTheDocument();
  });

  it('should display message when workspace not found', async () => {
    window.localStorage.removeItem('workspaceName');

    const {queryByText} = render(<Header />);

    expect(queryByText('Workspace')).not.toBeInTheDocument();
  });

  it("should display the user's name", async () => {
    const {findByText, queryByTestId} = render(<Header />);

    expect(queryByTestId('Account__loader')).toBeInTheDocument();

    expect(
      await findByText(`Hello, ${mockServer.getAccount().name}!`),
    ).toBeInTheDocument();
    expect(queryByTestId('Account__loader')).not.toBeInTheDocument();
  });

  it("should display user's email as a fallback", async () => {
    setIdTokenForAccount({id: 'ff7', email: 'barret.wallace@avalanche.net'});

    const {findByText} = render(<Header />);

    expect(
      await findByText(`Hello, barret.wallace@avalanche.net!`),
    ).toBeInTheDocument();
  });
});
