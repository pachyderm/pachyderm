import {render, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import {
  click,
  MockDag,
  SUBSCRIPTION_TIMEOUT,
  withContextProviders,
} from '@dash-frontend/testHelpers';
import {
  lineageRoute,
  pipelineRoute,
  repoRoute,
} from '@dash-frontend/views/Project/utils/routes';

describe('useRouteController', () => {
  const projectId = '2';

  const TestBed = withContextProviders(() => <MockDag />);

  afterEach(() => {
    window.history.replaceState('', '', lineageRoute({projectId}));
  });

  it('should derive the correct selected repo from the url', async () => {
    window.history.replaceState('', '', lineageRoute({projectId}));
    window.history.replaceState(
      '',
      '',
      repoRoute({projectId, repoId: 'likelihoods', branchId: 'master'}),
    );

    const {findByText} = render(<TestBed />);

    expect(
      await findByText(
        'Selected node: likelihoods_repo',
        {},
        {timeout: SUBSCRIPTION_TIMEOUT},
      ),
    ).toBeInTheDocument();
  });

  it('should derive the correct selected pipeline from the url', async () => {
    window.history.replaceState('', '', lineageRoute({projectId}));
    window.history.replaceState(
      '',
      '',
      pipelineRoute({projectId, pipelineId: 'likelihoods'}),
    );

    const {findByText} = render(<TestBed />);

    expect(
      await findByText(
        'Selected node: likelihoods',
        {},
        {timeout: SUBSCRIPTION_TIMEOUT},
      ),
    ).toBeInTheDocument();
  });

  it('should update the url correctly when selecting a repo', async () => {
    window.history.replaceState('', '', lineageRoute({projectId}));
    window.history.replaceState(
      '',
      '',
      pipelineRoute({projectId, pipelineId: 'likelihoods'}),
    );

    const {findByText} = render(<TestBed />);

    const imagesRepo = await findByText(
      '2 node id: likelihoods_repo',
      {},
      {timeout: SUBSCRIPTION_TIMEOUT},
    );

    await click(imagesRepo);

    await waitFor(() =>
      expect(window.location.pathname).toBe(
        repoRoute({projectId, repoId: 'likelihoods', branchId: 'master'}),
      ),
    );
  });

  it('should update the url correctly when selecting a pipeline', async () => {
    window.history.replaceState('', '', lineageRoute({projectId}));
    window.history.replaceState(
      '',
      '',
      repoRoute({projectId, repoId: 'likelihoods', branchId: 'master'}),
    );

    const {findByText} = render(<TestBed />);

    const likelihoodsPipeline = await findByText(
      '1 node id: likelihoods',
      {},
      {timeout: SUBSCRIPTION_TIMEOUT},
    );

    await click(likelihoodsPipeline);

    await waitFor(() =>
      expect(window.location.pathname).toBe(
        pipelineRoute({projectId, pipelineId: 'likelihoods'}),
      ),
    );
  });

  it('should not update the url when selecting an egress node', async () => {
    window.history.replaceState('', '', lineageRoute({projectId: '5'}));

    const {findByText} = render(<TestBed />);

    const egress = await findByText(
      '5 node id: https://egress.com',
      {},
      {timeout: SUBSCRIPTION_TIMEOUT},
    );

    await click(egress);

    await waitFor(() =>
      expect(window.location.pathname).toBe(lineageRoute({projectId: '5'})),
    );
  });
});
