import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {account} from '@dash-frontend/api/auth';
import {EMAIL_SUPPORT, SLACK_SUPPORT} from '@dash-frontend/constants/links';
import {
  mockGetVersionInfo,
  mockGetEnterpriseInfo,
  mockGetEnterpriseInfoInactive,
} from '@dash-frontend/mocks';
import {
  withContextProviders,
  click,
  loginUser,
} from '@dash-frontend/testHelpers';

import HeaderDropdownComponent from '../HeaderDropdown';

jest.mock('@dash-frontend/api/auth', () => ({
  account: jest.fn(() => ({
    id: 'unauthenticated',
    email: '',
    name: 'User',
  })),
}));

describe('HeaderDropdown', () => {
  const server = setupServer();

  beforeAll(() => server.listen());

  beforeEach(() => {
    window.localStorage.clear();
    server.resetHandlers();
    server.use(mockGetVersionInfo());
    server.use(mockGetEnterpriseInfo());
  });

  afterAll(() => server.close());

  const HeaderDropdown = withContextProviders(() => {
    return <HeaderDropdownComponent />;
  });

  describe('with Auth', () => {
    beforeEach(() => {
      loginUser();
    });

    it('should show account name, versions, and copyright', async () => {
      jest.mocked(account).mockResolvedValue({
        id: '1234567890',
        email: '',
        name: 'User Test',
      });

      render(<HeaderDropdown />);

      await click(
        screen.getByRole('button', {
          name: /header menu/i,
        }),
      );

      expect(await screen.findByText('User Test')).toBeInTheDocument();
      expect(await screen.findByText('Console test')).toBeInTheDocument();
      expect(screen.getByText('Pachd 0.0.0')).toBeInTheDocument();
    });

    it("should display the user's email as a fallback", async () => {
      jest.mocked(account).mockResolvedValue({
        id: '1234567890',
        email: 'email@user.com',
        name: '',
      });

      render(<HeaderDropdown />);

      await click(
        screen.getByRole('button', {
          name: /header menu/i,
        }),
      );

      expect(await screen.findByText('email@user.com')).toBeInTheDocument();
    });

    it('should direct users to email if enterprise is active', async () => {
      window.open = jest.fn();

      render(<HeaderDropdown />);

      await click(
        screen.getByRole('button', {
          name: /header menu/i,
        }),
      );
      await click(await screen.findByRole('menuitem', {name: 'email support'}));

      expect(window.open).toHaveBeenCalledTimes(1);
      expect(window.open).toHaveBeenCalledWith(
        expect.stringMatching(new RegExp(EMAIL_SUPPORT)),
      );
    });
  });

  describe('without Auth', () => {
    beforeEach(() => {
      server.resetHandlers();
      server.use(mockGetVersionInfo());
      server.use(mockGetEnterpriseInfoInactive());
    });

    it('should direct users to slack if enterprise is inactive', async () => {
      window.open = jest.fn();

      render(<HeaderDropdown />);
      await click(
        screen.getByRole('button', {
          name: /header menu/i,
        }),
      );
      await click(screen.getByRole('menuitem', {name: 'open slack support'}));

      expect(window.open).toHaveBeenCalledTimes(1);
      expect(window.open).toHaveBeenCalledWith(SLACK_SUPPORT);
    });
  });
});
