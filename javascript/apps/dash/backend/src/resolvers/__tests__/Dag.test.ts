import {executeOperation} from '@dash-backend/testHelpers';
import {Dag} from '@graphqlTypes';

const doesLinkExistInDag = (
  expectedLink: {source: string; target: string},
  dag: Dag | undefined,
) => {
  if (!dag) {
    return false;
  }

  const sourceNodeIndex = dag.nodes.findIndex(
    (node) => node.name === expectedLink.source,
  );
  const targetNodeIndex = dag.nodes.findIndex(
    (node) => node.name === expectedLink.target,
  );

  return dag.links.some((link) => {
    return link.source === sourceNodeIndex && link.target === targetNodeIndex;
  });
};

describe('Dag resolver', () => {
  it('should resolve dag data', async () => {
    const {data} = await executeOperation<{dag: Dag}>('getDag', {
      args: {projectId: '1'},
    });

    const dag = data?.dag;

    expect(dag?.links.length).toBe(5);
    expect(
      doesLinkExistInDag({source: 'montage', target: 'montage_repo'}, dag),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'edges', target: 'edges_repo'}, dag),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'edges_repo', target: 'montage'}, dag),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'images_repo', target: 'montage'}, dag),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'images_repo', target: 'edges'}, dag),
    ).toBe(true);
  });

  it('should resolve disconnected components of a dag', async () => {
    const {data} = await executeOperation<{dags: Dag[]}>('getDags', {
      args: {projectId: '2'},
    });

    const dags = data?.dags;

    expect(dags?.length).toBe(3);

    expect(dags?.[0].links.length).toBe(6);
    expect(
      doesLinkExistInDag(
        {source: 'samples_repo', target: 'likelihoods'},
        dags?.[0],
      ),
    ).toBe(true);
    expect(
      doesLinkExistInDag(
        {source: 'reference_repo', target: 'likelihoods'},
        dags?.[0],
      ),
    ).toBe(true);
    expect(
      doesLinkExistInDag(
        {source: 'reference_repo', target: 'joint_call'},
        dags?.[0],
      ),
    ).toBe(true);
    expect(
      doesLinkExistInDag(
        {source: 'likelihoods', target: 'likelihoods_repo'},
        dags?.[0],
      ),
    ).toBe(true);
    expect(
      doesLinkExistInDag(
        {source: 'joint_call', target: 'joint_call_repo'},
        dags?.[0],
      ),
    ).toBe(true);
    expect(
      doesLinkExistInDag(
        {source: 'likelihoods_repo', target: 'joint_call'},
        dags?.[0],
      ),
    ).toBe(true);

    expect(dags?.[1].links.length).toBe(2);
    expect(
      doesLinkExistInDag(
        {source: 'training_repo', target: 'models'},
        dags?.[1],
      ),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'models', target: 'models_repo'}, dags?.[1]),
    ).toBe(true);

    expect(dags?.[2].links.length).toBe(14);
    expect(
      doesLinkExistInDag({source: 'raw_data_repo', target: 'split'}, dags?.[2]),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'split', target: 'split_repo'}, dags?.[2]),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'split_repo', target: 'model'}, dags?.[2]),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'model', target: 'model_repo'}, dags?.[2]),
    ).toBe(true);
    expect(
      doesLinkExistInDag(
        {source: 'parameters_repo', target: 'model'},
        dags?.[2],
      ),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'model_repo', target: 'test'}, dags?.[2]),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'split_repo', target: 'test'}, dags?.[2]),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'test', target: 'test_repo'}, dags?.[2]),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'test_repo', target: 'select'}, dags?.[2]),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'model_repo', target: 'select'}, dags?.[2]),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'select', target: 'select_repo'}, dags?.[2]),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'model_repo', target: 'detect'}, dags?.[2]),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'images_repo', target: 'detect'}, dags?.[2]),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'detect', target: 'detect_repo'}, dags?.[2]),
    ).toBe(true);
  });
});
