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
});
