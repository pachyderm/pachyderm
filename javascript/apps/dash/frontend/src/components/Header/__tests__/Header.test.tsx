import {render, waitFor} from '@testing-library/react';
import React from 'react';

import HeaderComponent from '@dash-frontend/components/Header';
import {
  mockServer,
  setIdTokenForAccount,
  withContextProviders,
} from '@dash-frontend/testHelpers';

describe('Header', () => {
  const Header = withContextProviders(HeaderComponent);

  describe('landing header', () => {
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

  describe('project header', () => {
    it('should display notification badge if the project has unhealthy jobs', async () => {
      window.history.replaceState('', '', '/project/2');

      const {findByLabelText} = render(<Header />);

      expect(await findByLabelText('Number of failed jobs')).toHaveTextContent(
        '1',
      );
    });

    it('should not display notification badge for projects with no unhealthy jobs', async () => {
      window.history.replaceState('', '', '/project/3');

      const {queryByLabelText, queryByTestId} = render(<Header />);

      await waitFor(() =>
        expect(queryByTestId('ProjectHeader__projectNameLoader')).toBeNull(),
      );
      expect(queryByLabelText('Number of failed')).not.toBeInTheDocument();
    });
  });
});
