import {mockPipelinesQuery} from '@graphqlTypes';
import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockGetEnterpriseInfo,
  mockGetEnterpriseInfoInactive,
  mockGetEnterpriseInfoExpiring,
  mockGetManyPipelinesWithManyWorkers,
  getFutureTimeStamp,
  buildPipeline,
} from '@dash-frontend/mocks';
import {withContextProviders} from '@dash-frontend/testHelpers';

import CommunityEditionBannerComponent from '../';
import {PIPELINE_LIMIT} from '../useCommunityEditionBanner';

describe('CommunityEditionBanner', () => {
  const server = setupServer();
  const CommunityEditionBanner = withContextProviders<
    typeof CommunityEditionBannerComponent
  >((props) => {
    return <CommunityEditionBannerComponent {...props} />;
  });

  beforeAll(() => server.listen());

  afterAll(() => server.close());

  beforeEach(() => {
    window.history.replaceState({}, '', '/project/default/');
    server.resetHandlers();
    server.use(mockGetManyPipelinesWithManyWorkers());
    server.use(mockGetEnterpriseInfoInactive());
  });

  it('should show the banner when enterprise is inactive', async () => {
    render(<CommunityEditionBanner />);
    expect(await screen.findByText('Community Edition')).toBeInTheDocument();
    expect(
      await screen.findByText('Upgrade to Enterprise'),
    ).toBeInTheDocument();
  });

  it('should show the banner when enterprise is expiring', async () => {
    server.use(mockGetEnterpriseInfoExpiring());

    const expiration = getFutureTimeStamp(1);

    render(<CommunityEditionBanner expiration={expiration} />);

    expect(await screen.findByText('Enterprise Key')).toBeInTheDocument();
    expect(
      await screen.findByText('Access ends in 24 hours'),
    ).toBeInTheDocument();
  });

  it('should not show the banner when enterprise is active', async () => {
    server.use(mockGetEnterpriseInfo());

    const expiration = getFutureTimeStamp();

    render(<CommunityEditionBanner expiration={expiration} />);
    expect(screen.queryByText('Community Edition')).not.toBeInTheDocument();
  });

  it('should show a warning when pipeline limit is reached', async () => {
    server.use(mockGetManyPipelinesWithManyWorkers(16));

    render(<CommunityEditionBanner />);
    expect(
      await screen.findByText('Reaching pipeline limit'),
    ).toBeInTheDocument();
  });

  it('should show a warning when worker limit is reached', async () => {
    server.use(mockGetManyPipelinesWithManyWorkers(1, 8));

    render(<CommunityEditionBanner />);
    expect(
      await screen.findByText('Reaching worker limit (8 per pipeline)'),
    ).toBeInTheDocument();
  });

  it('should show total pipelines across projects', async () => {
    server.use(
      mockPipelinesQuery((req, res, ctx) => {
        // Component should always be sending an empty array of project ids
        if (req.variables.args.projectIds.length > 0) {
          res(ctx.data({pipelines: []}));
        }

        return res(
          ctx.data({pipelines: [buildPipeline({}), buildPipeline({})]}),
        );
      }),
    );

    render(<CommunityEditionBanner />);
    expect(
      await screen.findByText(`Pipelines: 2/${PIPELINE_LIMIT}`),
    ).toBeInTheDocument();
  });
});
