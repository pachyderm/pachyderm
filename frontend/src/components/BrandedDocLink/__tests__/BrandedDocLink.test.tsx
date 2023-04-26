import enterpriseStates from '@dash-backend/mock/fixtures/enterprise';
import {render, screen, waitFor} from '@testing-library/react';
import React from 'react';

import {withContextProviders, mockServer} from '@dash-frontend/testHelpers';

import BrandedDocLinkComponent from '../BrandedDocLink';

describe('BrandedDocLink', () => {
  const BrandedDocLink = withContextProviders<typeof BrandedDocLinkComponent>(
    (props) => (
      <BrandedDocLinkComponent {...props}>click here</BrandedDocLinkComponent>
    ),
  );

  beforeEach(() => {
    window.history.replaceState({}, '', '/');
  });

  it('should show the pachyderm docs when enterprise is inactive', async () => {
    mockServer.getState().enterprise = enterpriseStates.inactive;

    render(<BrandedDocLink pathWithoutDomain="fruit" />);

    expect(
      screen.getByRole('link', {
        name: /click here/i,
      }),
    ).toHaveAttribute('href', 'https://docs.pachyderm.com/latest/fruit');
  });

  it('should show the HPE docs when enterprise is active', async () => {
    mockServer.getState().enterprise = enterpriseStates.active;

    render(<BrandedDocLink pathWithoutDomain="fruit" />);

    await waitFor(() =>
      expect(
        screen.getByRole('link', {
          name: /click here/i,
        }),
      ).toHaveAttribute('href', 'https://mldm.pachyderm.com/latest/fruit'),
    );
  });

  it('removes leading slashes from to path', async () => {
    mockServer.getState().enterprise = enterpriseStates.inactive;

    render(<BrandedDocLink pathWithoutDomain="/fruit" />);

    const link = await screen.findByRole('link', {
      name: /click here/i,
    });

    expect(link).toHaveAttribute(
      'href',
      expect.not.stringContaining('//fruit'),
    );
    expect(link).toHaveAttribute('href', expect.stringContaining('/fruit'));
  });
});
