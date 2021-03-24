import {render} from '@testing-library/react';
import React from 'react';

import HeaderComponent from '@dash-frontend/components/Header';
import {
  mockServer,
  setIdTokenForAccount,
  withContextProviders,
} from '@dash-frontend/testHelpers';

describe('Header', () => {
  const Header = withContextProviders(HeaderComponent);

  it("should display the user's name", async () => {
    const {findByText, queryByTestId} = render(<Header />);

    expect(queryByTestId('Account__loader')).toBeInTheDocument();

    expect(
      await findByText(`Hello, ${mockServer.state.account.name}!`),
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
