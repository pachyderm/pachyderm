import enterpriseStates from '@dash-backend/mock/fixtures/enterprise';
import {render, screen, waitFor} from '@testing-library/react';
import React from 'react';

import {EMAIL_SUPPORT, SLACK_SUPPORT} from '@dash-frontend/constants/links';
import {withContextProviders, mockServer} from '@dash-frontend/testHelpers';

import ErrorStateSupportLinkComponent from '../ErrorStateSupportLink';

describe('ErrorStateSupportLink', () => {
  const ErrorStateSupportLink = withContextProviders(() => {
    return <ErrorStateSupportLinkComponent title="title" message="message" />;
  });

  beforeEach(() => {
    window.history.replaceState({}, '', '/');
  });

  it('should show a slack link if enterprise is inactive', async () => {
    mockServer.getState().enterprise = enterpriseStates.inactive;

    render(<ErrorStateSupportLink />);
    const link = screen.getByRole('link');
    expect(link).toHaveAttribute('href', SLACK_SUPPORT);
  });

  it('should show an email link if enterprise is active', async () => {
    mockServer.getState().enterprise = enterpriseStates.active;

    render(<ErrorStateSupportLink />);
    const link = screen.getByRole('link');

    await waitFor(() => expect(link).toHaveAttribute('href', EMAIL_SUPPORT));
  });
});
