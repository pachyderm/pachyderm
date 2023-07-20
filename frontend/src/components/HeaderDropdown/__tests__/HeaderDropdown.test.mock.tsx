import enterpriseStates from '@dash-backend/mock/fixtures/enterprise';
import {render, screen} from '@testing-library/react';
import React from 'react';

import {EMAIL_SUPPORT, SLACK_SUPPORT} from '@dash-frontend/constants/links';
import {
  withContextProviders,
  mockServer,
  click,
  setIdTokenForAccount,
} from '@dash-frontend/testHelpers';

import HeaderDropdownComponent from '../HeaderDropdown';

describe('HeaderDropdown', () => {
  const HeaderDropdown = withContextProviders(() => {
    return <HeaderDropdownComponent />;
  });

  it('should show account name, versions, and copyright', async () => {
    render(<HeaderDropdown />);
    await click(
      screen.getByRole('button', {
        name: /header menu/i,
      }),
    );

    expect(
      await screen.findByText(`${mockServer.getAccount().name}`),
    ).toBeInTheDocument();
    expect(await screen.findByText(/development/)).toBeInTheDocument();
    expect(await screen.findByText(/2\.5\.4additional/)).toBeInTheDocument();
    expect(await screen.findByText(/, HPE/)).toBeInTheDocument();
  });

  it("should display the user's email as a fallback", async () => {
    setIdTokenForAccount({id: 'ff7', email: 'barret.wallace@avalanche.net'});

    render(<HeaderDropdown />);

    expect(
      await screen.findByText('barret.wallace@avalanche.net'),
    ).toBeInTheDocument();
  });

  it('should direct users to slack if enterprise is inactive', async () => {
    window.open = jest.fn();
    mockServer.getState().enterprise = enterpriseStates.inactive;

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

  it('should direct users to email if enterprise is active', async () => {
    window.open = jest.fn();
    mockServer.getState().enterprise = enterpriseStates.active;

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

  it('should give users a pachctl command to set their active project', async () => {
    window.history.replaceState('', '', `/lineage/testProject`);

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
      await screen.findByText('Set Active Project: "testProject"'),
    ).toBeInTheDocument();
    await click(screen.getByRole('button', {name: 'Copy'}));

    expect(window.document.execCommand).toHaveBeenCalledWith('copy');
  });
});
