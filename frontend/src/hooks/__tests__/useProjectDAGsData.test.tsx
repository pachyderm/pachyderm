import {render, waitForElementToBeRemoved} from '@testing-library/react';
import React from 'react';
import {Route} from 'react-router';

import {DagDirection} from '@dash-frontend/lib/types';
import {withContextProviders} from '@dash-frontend/testHelpers';

import {useProjectDagsData} from '../useProjectDAGsData';
import useUrlState from '../useUrlState';

const ProjectsComponent = () => {
  const {projectId, jobId} = useUrlState();
  const {dags, loading} = useProjectDagsData({
    projectId,
    nodeHeight: 60,
    nodeWidth: 120,
    direction: DagDirection.RIGHT,
    jobSetId: jobId,
  });

  if (loading) return <span>Loading</span>;

  return (
    <div>
      {(dags || []).map((dag, i) => {
        return (
          <div key={i}>
            <div>
              {i} id: {dag.id}
            </div>
            {dag.nodes.map((node, i) => {
              return (
                <div key={node.id}>
                  <span>
                    {i} node id: {node.id}
                  </span>
                  <span>
                    {i} node name: {node.name}
                  </span>
                  <span>
                    {i} node type: {node.type}
                  </span>
                  <span>
                    {i} node state: {node.state}
                  </span>
                </div>
              );
            })}
            {dag.links.map((link, i) => {
              return (
                <div key={link.id}>
                  <span>
                    {i} link source: {link.source}
                  </span>
                  <span>
                    {i} link target: {link.target}
                  </span>
                  <span>
                    {i} link state: {link.state}
                  </span>
                </div>
              );
            })}
          </div>
        );
      })}
    </div>
  );
};

const TestBed = withContextProviders(() => {
  return (
    <Route path="/project/:projectId">
      <ProjectsComponent />
    </Route>
  );
});

describe('useProjects', () => {
  it('should get dag data', async () => {
    window.history.replaceState('', '', '/project/1');

    const {findByText} = render(<TestBed />);

    await waitForElementToBeRemoved(await findByText('Loading'), {
      timeout: 10000,
    });

    const node0Id = await findByText('0 node id: montage_repo');
    const node1Id = await findByText('1 node id: montage');
    const node2Id = await findByText('2 node id: edges_repo');
    const node3Id = await findByText('3 node id: edges');
    const node4Id = await findByText('4 node id: images_repo');

    expect(node0Id).toBeInTheDocument();
    expect(node1Id).toBeInTheDocument();
    expect(node2Id).toBeInTheDocument();
    expect(node3Id).toBeInTheDocument();
    expect(node4Id).toBeInTheDocument();

    const node0name = await findByText('0 node name: montage_repo');
    const node0type = await findByText('0 node type: OUTPUT_REPO');
    const node0state = await findByText('0 node state:');
    const node1state = await findByText('1 node state: ERROR');

    expect(node0name).toBeInTheDocument();
    expect(node0type).toBeInTheDocument();
    expect(node0state).toBeInTheDocument();
    expect(node1state).toBeInTheDocument();

    const link0Source = await findByText('0 link source: montage');
    const link1Source = await findByText('1 link source: edges');
    const link2Source = await findByText('2 link source: images_repo');
    const link3Source = await findByText('3 link source: images_repo');
    const link4Source = await findByText('4 link source: edges_repo');

    expect(link0Source).toBeInTheDocument();
    expect(link1Source).toBeInTheDocument();
    expect(link2Source).toBeInTheDocument();
    expect(link3Source).toBeInTheDocument();
    expect(link4Source).toBeInTheDocument();

    const link0Target = await findByText('0 link target: montage_repo');
    const link1Target = await findByText('1 link target: edges_repo');
    const link2Target = await findByText('2 link target: montage');
    const link3Target = await findByText('3 link target: edges');
    const link4Target = await findByText('4 link target: montage');

    expect(link0Target).toBeInTheDocument();
    expect(link1Target).toBeInTheDocument();
    expect(link2Target).toBeInTheDocument();
    expect(link3Target).toBeInTheDocument();
    expect(link4Target).toBeInTheDocument();

    const link0state = await findByText('0 link state: JOB_CREATED');
    const link1state = await findByText('1 link state: JOB_CREATED');

    expect(link0state).toBeInTheDocument();
    expect(link1state).toBeInTheDocument();
  });

  it('should correctly render cron inputs', async () => {
    window.history.replaceState('', '', '/project/3');

    const {findByText} = render(<TestBed />);

    await waitForElementToBeRemoved(await findByText('Loading'), {
      timeout: 10000,
    });

    const cronLinkSource = await findByText('0 link source: cron_repo');
    const cronLinkTarget = await findByText('0 link target: processor');
    const processorLinkSource = await findByText('1 link source: processor');
    const processorLinkTarget = await findByText(
      '1 link target: processor_repo',
    );

    expect(cronLinkSource).toBeInTheDocument();
    expect(cronLinkTarget).toBeInTheDocument();
    expect(processorLinkSource).toBeInTheDocument();
    expect(processorLinkTarget).toBeInTheDocument();
  });

  it('should send dag id as name of oldest repo', async () => {
    window.history.replaceState('', '', '/project/2');

    const {findByText} = render(<TestBed />);

    await waitForElementToBeRemoved(await findByText('Loading'), {
      timeout: 10000,
    });

    const id = await findByText('0 id: samples_repo');
    expect(id).toBeInTheDocument();
  });

  it('should send dag id as jobset id for job sub-dags', async () => {
    window.history.replaceState(
      '',
      '',
      '/project/1/jobs/33b9af7d5d4343219bc8e02ff44cd55a',
    );

    const {findByText} = render(<TestBed />);

    await waitForElementToBeRemoved(await findByText('Loading'), {
      timeout: 10000,
    });

    const id = await findByText('0 id: 33b9af7d5d4343219bc8e02ff44cd55a');
    expect(id).toBeInTheDocument();
  });
});
