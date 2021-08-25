import {render} from '@testing-library/react';
import React from 'react';

import {
  withContextProviders,
  mockServer,
  setIdTokenForAccount,
  click,
} from '@dash-frontend/testHelpers';

import LandingHeader from '../../LandingHeader';

describe('LandingHeader', () => {
  const Header = withContextProviders(LandingHeader);

  afterEach(() => {
    localStorage.removeItem('workspaceName');
    localStorage.removeItem('pachVersion');
    localStorage.removeItem(
      'grpcs://grpc-hub-c0-j95orhflud.clusters.pachyderm.io:31400',
    );
  });

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

  it('should allow the user to access connection information for their workspace', async () => {
    const workspaceName = 'test';
    const pachVersion = '2.0';
    const pachdAddress =
      'grpcs://grpc-hub-c0-j95orhflud.clusters.pachyderm.io:31400';

    localStorage.setItem('workspaceName', workspaceName);
    localStorage.setItem('pachVersion', pachVersion);
    localStorage.setItem('pachdAddress', pachdAddress);

    const {findByText, findByTestId} = render(<Header />);

    const connectToWorkspaceButton = await findByText('Connect to Workspace');

    click(connectToWorkspaceButton);

    expect(
      await findByText(
        `echo '{"pachd_address": "${pachdAddress}", "source": 2}' | pachctl config set context "${workspaceName}" --overwrite && pachctl config set active-context "${workspaceName}"`,
      ),
    ).toBeInTheDocument();
    expect(
      await findByTestId('ConnectModal__pachctlVersion'),
    ).toHaveTextContent(pachVersion);
    expect(await findByTestId('ConnectModal__pachdVersion')).toHaveTextContent(
      pachVersion,
    );
  });
});
