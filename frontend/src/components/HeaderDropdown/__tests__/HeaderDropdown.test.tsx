import {mockGetAccountQuery} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {EMAIL_SUPPORT, SLACK_SUPPORT} from '@dash-frontend/constants/links';
import {
  mockGetVersionInfo,
  mockGetAccountAuth,
  mockGetEnterpriseInfo,
} from '@dash-frontend/mocks';
import {
  withContextProviders,
  click,
  loginUser,
} from '@dash-frontend/testHelpers';

import HeaderDropdownComponent from '../HeaderDropdown';

describe('HeaderDropdown', () => {
  const server = setupServer();

  beforeAll(() => {
    server.listen();
  });

  afterAll(() => server.close());

  const HeaderDropdown = withContextProviders(() => {
    return <HeaderDropdownComponent />;
  });

  describe('with Auth', () => {
    beforeAll(() => {
      server.use(mockGetVersionInfo());
      server.use(mockGetAccountAuth());
      server.use(mockGetEnterpriseInfo());
      loginUser();
    });

    afterAll(() => {
      server.resetHandlers();
      window.localStorage.clear();
    });

    it('should show account name, versions, and copyright', async () => {
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
      server.use(
        mockGetAccountQuery((_req, res, ctx) => {
          return res(
            ctx.data({
              account: {
                email: 'email@user.com',
                id: 'TestUsername',
                name: '',
              },
            }),
          );
        }),
      );

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
    beforeAll(() => {
      server.use(mockGetVersionInfo());
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

  it('should give users a pachctl command to set their active project', async () => {
    window.history.replaceState('', '', `/lineage/projectA`);

    render(<HeaderDropdown />);

    await click(
      screen.getByRole('button', {
        name: /header menu/i,
      }),
    );
    await click(
      await screen.findByRole('menuitem', {
        name: /set active project/i,
      }),
    );

    expect(
      await screen.findByText('Set Active Project: "projectA"'),
    ).toBeInTheDocument();
    await click(screen.getByRole('button', {name: 'Copy'}));
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith(
      'pachctl config update context --project projectA',
    );
  });
});
