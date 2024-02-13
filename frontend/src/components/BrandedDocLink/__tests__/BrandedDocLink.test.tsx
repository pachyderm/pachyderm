import {render, screen, waitFor} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockGetVersionInfo,
  mockGetEnterpriseInfo,
  mockGetEnterpriseInfoInactive,
} from '@dash-frontend/mocks';
import {withContextProviders, loginUser} from '@dash-frontend/testHelpers';

import BrandedDocLinkComponent from '../BrandedDocLink';

describe('BrandedDocLink', () => {
  const server = setupServer();

  const BrandedDocLink = withContextProviders<typeof BrandedDocLinkComponent>(
    (props) => (
      <BrandedDocLinkComponent {...props}>click here</BrandedDocLinkComponent>
    ),
  );

  beforeAll(() => server.listen());

  beforeEach(() => {
    window.localStorage.clear();
    server.resetHandlers();
    server.use(mockGetVersionInfo());
    server.use(mockGetEnterpriseInfoInactive());
  });

  afterAll(() => server.close());

  it('should show the pachyderm docs when enterprise is inactive', async () => {
    render(<BrandedDocLink pathWithoutDomain="fruit" />);

    await waitFor(() =>
      expect(
        screen.getByRole('link', {
          name: /click here/i,
        }),
      ).toHaveAttribute('href', 'https://docs.pachyderm.com/0.0.x/fruit'),
    );
  });

  it('removes leading slashes from to path', async () => {
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

  it('should show the HPE docs when enterprise is active', async () => {
    server.use(mockGetEnterpriseInfo());
    loginUser();

    render(<BrandedDocLink pathWithoutDomain="fruit" />);

    await waitFor(() =>
      expect(
        screen.getByRole('link', {
          name: /click here/i,
        }),
      ).toHaveAttribute('href', 'https://mldm.pachyderm.com/0.0.x/fruit'),
    );
  });
});
