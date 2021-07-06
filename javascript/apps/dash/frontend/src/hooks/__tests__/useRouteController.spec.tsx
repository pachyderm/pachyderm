import {render, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';
import {MockDAG, withContextProviders} from '@dash-frontend/testHelpers';
import {
  NODE_HEIGHT,
  NODE_WIDTH,
} from '@dash-frontend/views/Project/constants/nodeSizes';
import {DagDirection} from '@graphqlTypes';

describe('useRouteController', () => {
  const projectId = '2';

  const TestBed = withContextProviders(({id = '2'}: {id: string}) => {
    const {dags = [], loading} = useProjectDagsData({
      projectId: id,
      nodeHeight: NODE_HEIGHT,
      nodeWidth: NODE_WIDTH,
      direction: DagDirection.RIGHT,
    });

    if (loading) {
      return <>Loading...</>;
    }

    return <MockDAG dag={dags[0]} />;
  });

  afterEach(() => {
    window.history.replaceState('', '', `/project/${projectId}`);
  });

  it('should derive the correct selected repo from the url', async () => {
    window.history.replaceState(
      '',
      '',
      `/project/${projectId}/repo/likelihoods/branch/master`,
    );

    const {findByText} = render(<TestBed />);

    expect(
      await findByText('Selected node: likelihoods_repo', {}, {timeout: 10000}),
    );
  });

  it('should derive the correct selected pipeline from the url', async () => {
    window.history.replaceState(
      '',
      '',
      `/project/${projectId}/pipeline/likelihoods`,
    );

    const {findByText} = render(<TestBed />);

    expect(
      await findByText('Selected node: likelihoods', {}, {timeout: 10000}),
    );
  });

  it('should update the url correctly when selecting a repo', async () => {
    window.history.replaceState(
      '',
      '',
      `/project/${projectId}/pipeline/likelihoods`,
    );

    const {findByText} = render(<TestBed />);

    const imagesRepo = await findByText(
      'likelihoods_repo',
      {},
      {timeout: 10000},
    );

    // TODO: replace with event helpers
    userEvent.click(imagesRepo);

    await waitFor(() =>
      expect(window.location.pathname).toBe(
        `/project/${projectId}/repo/likelihoods/branch/master`,
      ),
    );
  });

  it('should update the url correctly when selecting a pipeline', async () => {
    window.history.replaceState(
      '',
      '',
      `/project/${projectId}/repo/likelihoods/branch/master`,
    );

    const {findByText} = render(<TestBed />);

    const edgesPipeline = await findByText('likelihoods', {}, {timeout: 10000});

    // TODO: replace with event helpers
    userEvent.click(edgesPipeline);

    await waitFor(() =>
      expect(window.location.pathname).toBe(
        `/project/${projectId}/pipeline/likelihoods`,
      ),
    );
  });

  it('should not update the url when selecting an egress node', async () => {
    window.history.replaceState('', '', '/project/5');

    const {findByText} = render(<TestBed id="5" />);

    const egress = await findByText('https://egress.com', {}, {timeout: 10000});

    userEvent.click(egress);

    await waitFor(() => expect(window.location.pathname).toBe('/project/5'));
  });
});
