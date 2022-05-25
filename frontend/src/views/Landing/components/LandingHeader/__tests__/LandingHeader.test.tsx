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

  it("should display the user's name", async () => {
    const {findAllByText} = render(<Header />);

    expect(
      await findAllByText(`Hello, ${mockServer.getAccount().name}!`),
    ).toHaveLength(2);
  });

  it("should display user's email as a fallback", async () => {
    setIdTokenForAccount({id: 'ff7', email: 'barret.wallace@avalanche.net'});

    const {findAllByText} = render(<Header />);

    expect(
      await findAllByText(`Hello, barret.wallace@avalanche.net!`),
    ).toHaveLength(2);
  });
});
