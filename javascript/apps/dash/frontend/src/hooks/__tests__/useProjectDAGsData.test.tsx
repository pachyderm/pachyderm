import {render} from '@testing-library/react';
import React from 'react';

import {withContextProviders} from '@dash-frontend/testHelpers';
import {DagDirection} from '@graphqlTypes';

import {useProjectDagsData} from '../useProjectDAGsData';

const ProjectsComponent = withContextProviders(() => {
  const {dags, loading} = useProjectDagsData({
    projectId: '1',
    nodeHeight: 60,
    nodeWidth: 120,
    direction: DagDirection.RIGHT,
  });

  if (loading) return <span>Loading</span>;

  return (
    <div>
      {(dags || []).map((dag, i) => {
        return (
          <div key={i}>
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
                    {i} link id: {link.id}
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
});

describe('useProjects', () => {
  it('should get dag data', async () => {
    const {findByText} = render(<ProjectsComponent />);

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
    const node1state = await findByText('1 node state: PIPELINE_FAILURE');

    expect(node0name).toBeInTheDocument();
    expect(node0type).toBeInTheDocument();
    expect(node0state).toBeInTheDocument();
    expect(node1state).toBeInTheDocument();

    const link0id = await findByText('0 link id: montage-montage_repo');
    const link1id = await findByText('1 link id: edges-edges_repo');
    const link2id = await findByText('2 link id: images_repo-montage');
    const link3id = await findByText('3 link id: images_repo-edges');
    const link4id = await findByText('4 link id: edges_repo-montage');

    expect(link0id).toBeInTheDocument();
    expect(link1id).toBeInTheDocument();
    expect(link2id).toBeInTheDocument();
    expect(link3id).toBeInTheDocument();
    expect(link4id).toBeInTheDocument();

    const link0state = await findByText('0 link state: JOB_STARTING');
    const link1state = await findByText('1 link state: JOB_STARTING');

    expect(link0state).toBeInTheDocument();
    expect(link1state).toBeInTheDocument();
  });
});
