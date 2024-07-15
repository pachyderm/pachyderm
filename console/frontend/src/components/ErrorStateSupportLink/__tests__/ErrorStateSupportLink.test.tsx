import {render, screen, waitFor} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {EMAIL_SUPPORT, SLACK_SUPPORT} from '@dash-frontend/constants/links';
import {
  mockGetVersionInfo,
  mockGetEnterpriseInfo,
  mockGetEnterpriseInfoInactive,
} from '@dash-frontend/mocks';
import {withContextProviders, loginUser} from '@dash-frontend/testHelpers';

import ErrorStateSupportLinkComponent from '../ErrorStateSupportLink';

describe('ErrorStateSupportLink', () => {
  const server = setupServer();

  const ErrorStateSupportLink = withContextProviders(() => {
    return <ErrorStateSupportLinkComponent title="title" message="message" />;
  });

  beforeAll(() => server.listen());

  beforeEach(() => {
    window.localStorage.clear();
    server.resetHandlers();
    server.use(mockGetVersionInfo());
    server.use(mockGetEnterpriseInfoInactive());
  });

  afterAll(() => server.close());

  it('should show a slack link if enterprise is inactive', async () => {
    render(<ErrorStateSupportLink />);
    const link = screen.getByRole('link');
    expect(link).toHaveAttribute('href', SLACK_SUPPORT);
  });

  it('should show an email link if enterprise is active', async () => {
    server.use(mockGetEnterpriseInfo());
    loginUser();

    render(<ErrorStateSupportLink />);
    const link = screen.getByRole('link');

    await waitFor(() => expect(link).toHaveAttribute('href', EMAIL_SUPPORT));
  });
});
