import enterpriseStates from '@dash-backend/mock/fixtures/enterprise';
import {render} from '@testing-library/react';
import React from 'react';

import {withContextProviders, mockServer} from '@dash-frontend/testHelpers';

import CommunityEditionBannerProvider from '../CommunityEditionBannerProvider';

describe('CommunityEditionBanner', () => {
  const CommunityEditionBanner = withContextProviders(() => {
    return <CommunityEditionBannerProvider />;
  });

  beforeEach(() => {
    window.history.replaceState({}, '', '/');
  });

  it('should show the banner when enterprise is inactive', async () => {
    mockServer.getState().enterprise = enterpriseStates.inactive;

    const {findByText} = render(<CommunityEditionBanner />);
    expect(await findByText('Community Edition')).toBeInTheDocument();
  });

  it('should show the banner when enterprise is expiring', async () => {
    mockServer.getState().enterprise = enterpriseStates.expiring;

    const {findByText} = render(<CommunityEditionBanner />);

    expect(await findByText('Enterprise Key')).toBeInTheDocument();
    expect(await findByText('Access ends in 24 hours')).toBeInTheDocument();
  });

  it('should not show the banner when enterprise is active', async () => {
    mockServer.getState().enterprise = enterpriseStates.active;

    const {queryByText} = render(<CommunityEditionBanner />);
    expect(await queryByText('Community Edition')).not.toBeInTheDocument();
    expect(await queryByText('Enterprise Key')).not.toBeInTheDocument();
  });

  it('should show a warning when pipeline limit is reached', async () => {
    mockServer.getState().enterprise = enterpriseStates.inactive;
    window.history.replaceState({}, '', '/project/7/');

    const {findByText} = render(<CommunityEditionBanner />);
    expect(await findByText('Reaching pipeline limit')).toBeInTheDocument();
  });

  it('should show a warning when worker limit is reached', async () => {
    mockServer.getState().enterprise = enterpriseStates.inactive;
    window.history.replaceState({}, '', '/project/1/');

    const {findByText} = render(<CommunityEditionBanner />);
    expect(
      await findByText('Reaching worker limit (8 per pipeline)'),
    ).toBeInTheDocument();
  });
});
