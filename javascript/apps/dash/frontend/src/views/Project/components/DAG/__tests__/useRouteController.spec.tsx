import {render, waitFor} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import {useProjectDagsData} from '@dash-frontend/hooks/useProjectDAGsData';
import {MockDAG, withContextProviders} from '@dash-frontend/testHelpers';
import {
  NODE_HEIGHT,
  NODE_WIDTH,
} from '@dash-frontend/views/Project/constants/nodeSizes';

describe('useRouteController', () => {
  const projectId = '2';

  const TestBed = withContextProviders(() => {
    const {dags = [], loading} = useProjectDagsData({
      projectId,
      nodeHeight: NODE_HEIGHT,
      nodeWidth: NODE_WIDTH,
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
      `/project/${projectId}/dag/samples/repo/likelihoods/branch/master`,
    );

    const {findByText} = render(<TestBed />);

    expect(await findByText('Selected node: likelihoods_repo'));
  });

  it('should derive the correct selected pipeline from the url', async () => {
    window.history.replaceState(
      '',
      '',
      `/project/${projectId}/dag/samples/pipeline/likelihoods`,
    );

    const {findByText} = render(<TestBed />);

    expect(await findByText('Selected node: likelihoods'));
  });

  it('should update the url correctly when selecting a repo', async () => {
    window.history.replaceState(
      '',
      '',
      `/project/${projectId}/dag/samples/pipeline/likelihoods`,
    );

    const {findByText} = render(<TestBed />);

    const imagesRepo = await findByText('likelihoods_repo');

    // TODO: replace with event helpers
    userEvent.click(imagesRepo);

    await waitFor(() =>
      expect(window.location.pathname).toBe(
        `/project/${projectId}/dag/samples/repo/likelihoods/branch/master`,
      ),
    );
  });

  it('should update the url correctly when selecting a pipeline', async () => {
    window.history.replaceState(
      '',
      '',
      `/project/${projectId}/dag/samples/repo/likelihoods/branch/master`,
    );

    const {findByText} = render(<TestBed />);

    const edgesPipeline = await findByText('likelihoods');

    // TODO: replace with event helpers
    userEvent.click(edgesPipeline);

    await waitFor(() =>
      expect(window.location.pathname).toBe(
        `/project/${projectId}/dag/samples/pipeline/likelihoods`,
      ),
    );
  });
});
