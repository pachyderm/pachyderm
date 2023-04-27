import enterpriseStates from '@dash-backend/mock/fixtures/enterprise';
import {render, waitFor} from '@testing-library/react';
import React from 'react';

import {withContextProviders, mockServer} from '@dash-frontend/testHelpers';

import BrandedTitleComponent from '../BrandedTitle';

describe('BrandedDocLink', () => {
  const BrandedDocLink = withContextProviders<typeof BrandedTitleComponent>(
    (props) => <BrandedTitleComponent {...props} />,
  );

  it('should show the pachyderm title when enterprise is inactive', async () => {
    mockServer.getState().enterprise = enterpriseStates.inactive;

    render(<BrandedDocLink title="Project" />);

    await waitFor(() => {
      expect(document.title).toBe('Project - Pachyderm Console');
    });
  });

  it('should show the HPE title when enterprise is active', async () => {
    mockServer.getState().enterprise = enterpriseStates.active;

    render(<BrandedDocLink title="Project" />);

    await waitFor(() => {
      expect(document.title).toBe('Project - HPE MLDM');
    });
  });
});
