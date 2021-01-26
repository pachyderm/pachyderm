import {Dag} from 'generated/types';
import {query} from 'testHelpers';

const doesLinkExistInDag = (
  expectedLink: {source: string; target: string},
  dag: Dag,
) => {
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
    const {dag} = await query<{dag: Dag}>`
      query getDag {
        dag(args: {projectId: "tutorial"}) {
          nodes {
            name
            type
            error
            access
          }
          links {
            source
            target
            error
            active
          }
        }
      }
    `;

    expect(dag.links.length).toBe(5);
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
    const {dags} = await query<{dags: Dag[]}>`
      query getDags {
        dags(args: {projectId: "customerTeam"}) {
          nodes {
            name
            type
            error
            access
          }
          links {
            source
            target
            error
            active
          }
        }
      }
    `;

    expect(dags.length).toBe(2);

    expect(dags[0].links.length).toBe(6);
    expect(
      doesLinkExistInDag(
        {source: 'samples_repo', target: 'likelihoods'},
        dags[0],
      ),
    ).toBe(true);
    expect(
      doesLinkExistInDag(
        {source: 'reference_repo', target: 'likelihoods'},
        dags[0],
      ),
    ).toBe(true);
    expect(
      doesLinkExistInDag(
        {source: 'reference_repo', target: 'joint_call'},
        dags[0],
      ),
    ).toBe(true);
    expect(
      doesLinkExistInDag(
        {source: 'likelihoods', target: 'likelihoods_repo'},
        dags[0],
      ),
    ).toBe(true);
    expect(
      doesLinkExistInDag(
        {source: 'joint_call', target: 'joint_call_repo'},
        dags[0],
      ),
    ).toBe(true);
    expect(
      doesLinkExistInDag(
        {source: 'likelihoods_repo', target: 'joint_call'},
        dags[0],
      ),
    ).toBe(true);

    expect(dags[1].links.length).toBe(2);
    expect(
      doesLinkExistInDag({source: 'training_repo', target: 'models'}, dags[1]),
    ).toBe(true);
    expect(
      doesLinkExistInDag({source: 'models', target: 'models_repo'}, dags[1]),
    ).toBe(true);
  });
});
