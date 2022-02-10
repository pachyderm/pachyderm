import {NodeType, JobState, Vertex, Maybe} from '@graphqlTypes';
import ELK from 'elkjs/lib/elk.bundled.js';
import flatten from 'lodash/flatten';
import minBy from 'lodash/minBy';
import objectHash from 'object-hash';

import disconnectedComponents from './disconnectedComponents';
import {LinkInputData, NodeInputData, Link, Node, DagDirection} from './types';

interface DeriveTransferringStateOpts {
  targetNodeType: NodeType;
  targetNodeState?: Maybe<JobState>;
}

const deriveTransferringState = ({
  targetNodeType,
  targetNodeState,
}: DeriveTransferringStateOpts) => {
  // Should display transferring _from_ a corresponding egressing pipeline's output repo.
  // Because repos do not have a "state", we need to use the target node's state instead.
  return (
    (targetNodeType === NodeType.EGRESS &&
      targetNodeState === JobState.JOB_EGRESSING) ||
    (targetNodeType !== NodeType.EGRESS &&
      targetNodeState === JobState.JOB_RUNNING)
  );
};

const normalizeDAGData = async (
  vertices: Vertex[],
  nodeWidth: number,
  nodeHeight: number,
  direction: DagDirection = DagDirection.RIGHT,
  jobSetId?: string,
) => {
  const elk = new ELK();

  // Calculate the indicies of the nodes in the DAG, as
  // the "link" object requires the source & target attributes
  // to reference the index from the "nodes" list. This prevents
  // us from needing to run "indexOf" on each individual node
  // when building the list of links for the client.
  const correspondingIndex: {[key: string]: number} = vertices.reduce(
    (acc, val, index) => ({
      ...acc,
      [val.name]: index,
    }),
    {},
  );

  // create elk edges
  const elkEdges = flatten(
    vertices.map<LinkInputData[]>((node) => {
      if (node.type === NodeType.PIPELINE) {
        return node.parents.reduce<LinkInputData[]>((acc, parent) => {
          const parentName = `${parent}_repo`;
          const sourceIndex = correspondingIndex[parentName];
          const state = vertices[sourceIndex].jobState || undefined;

          return [
            ...acc,
            {
              id: objectHash({node: node, jobSetId, parent}),
              sources: [parentName],
              targets: [node.name],
              state,
              sourceState: vertices[sourceIndex].state || undefined,
              targetstate: node.state || undefined,
              sections: [],
              transferring: deriveTransferringState({
                targetNodeType: node.type,
                targetNodeState: node.jobState,
              }),
            },
          ];
        }, []);
      }

      return node.parents.reduce<LinkInputData[]>((acc, parentName) => {
        const sourceIndex = correspondingIndex[parentName || ''];
        const state = vertices[sourceIndex].jobState || undefined;

        return [
          ...acc,
          {
            // hashing here for consistency
            id: objectHash(`${parentName}-${node.name}-${jobSetId}`),
            sources: [parentName],
            targets: [node.name],
            state,
            sourceState: vertices[sourceIndex].state || undefined,
            targetstate: node.state || undefined,
            sections: [],
            transferring: deriveTransferringState({
              targetNodeType: vertices[sourceIndex].type,
              targetNodeState: state,
            }),
          },
        ];
      }, []);
    }),
  );

  // create elk children
  const elkChildren = vertices.map<NodeInputData>((node) => ({
    id: node.name,
    name: node.name,
    type: node.type,
    state: node.state ? node.state : undefined,
    access: node.access,
    x: 0,
    y: 0,
    width: nodeWidth,
    height: nodeHeight,
  }));

  const horizontal = direction === DagDirection.RIGHT;

  await elk.layout(
    {id: 'root', children: elkChildren, edges: elkEdges},
    {
      layoutOptions: {
        // the default value of '0' picks a psuedo random seed based
        // on system time, which makes the algorithm non-deterministic
        // https://www.eclipse.org/elk/reference/options/org-eclipse-elk-randomSeed.html
        // Elk doesn't, however, ignore order. Input nodes and edges need to be
        // based off stabalized (ordered) data (i.e. listRepo, listPipeline)
        'org.eclipse.elk.randomSeed': '1',
        'org.eclipse.elk.algorithm': 'disco',
        'org.eclipse.elk.aspectRatio': horizontal ? '0.2' : '10',
        'org.eclipse.elk.mergeEdges': 'false',
        'org.eclipse.elk.direction': direction,
        'org.eclipse.elk.layered.layering.strategy': 'INTERACTIVE',
        'org.eclipse.elk.disco.componentCompaction.componentLayoutAlgorithm':
          'layered',
        'spacing.componentComponent': horizontal ? '75' : '150',

        'org.eclipse.elk.layered.spacing.edgeNodeBetweenLayers': '20',
        'org.eclipse.elk.layered.spacing.nodeNodeBetweenLayers': '40',
      },
    },
  );

  // convert elk edges to graphl Link
  const links = elkEdges.map<Link>((edge) => {
    const {
      sections,
      sources,
      targets,
      junctionPoints,
      labels,
      layoutOptions,
      ...rest
    } = edge;

    return {
      ...rest,
      source: sources[0],
      target: targets[0],
      startPoint: sections?.[0].startPoint || {x: 0, y: 0},
      endPoint: sections?.[0].endPoint || {x: 0, y: 0},
      bendPoints: sections?.[0].bendPoints || [],
    };
  });

  //convert elk children to graphl Node
  const nodes = elkChildren.map<Node>((node) => {
    return {
      id: node.id,
      x: node.x || 0,
      y: node.y || 0,
      name: node.name,
      type: node.type,
      state: node.state,
      access: node.access,
    };
  });

  return {
    nodes,
    links,
  };
};

// find offset needed to horizontally or vertically align dag
const adjustDag = (
  {nodes, links}: {nodes: Node[]; links: Link[]},
  direction: DagDirection,
) => {
  const horizontal = direction === DagDirection.RIGHT;
  const xValues = nodes.map((n) => n.x);
  const yValues = nodes.map((n) => n.y);
  const minY = Math.min(...yValues);
  const minX = Math.min(...xValues);
  const offsetX = horizontal ? -minX : 0;
  const offsetY = !horizontal ? -minY : 0;

  const adjustedNodes = nodes.map((node) => {
    return {
      ...node,
      x: node.x + offsetX,
      y: node.y + offsetY,
    };
  });

  const adjustedLinks = links.map((link) => {
    return {
      ...link,
      startPoint: {
        x: link.startPoint.x + offsetX,
        y: link.startPoint.y + offsetY,
      },
      bendPoints: link.bendPoints.map((point) => ({
        x: point.x + offsetX,
        y: point.y + offsetY,
      })),
      endPoint: {
        x: link.endPoint.x + offsetX,
        y: link.endPoint.y + offsetY,
      },
    };
  });

  return {
    nodes: adjustedNodes,
    links: adjustedLinks,
  };
};

const buildDags = async (
  vertices: Vertex[],
  nodeWidth: number,
  nodeHeight: number,
  direction = DagDirection.RIGHT,
  jobSetId?: string,
) => {
  const {nodes, links} = await normalizeDAGData(
    vertices,
    nodeWidth,
    nodeHeight,
    direction,
    jobSetId,
  );
  return disconnectedComponents(nodes, links).map((component) => {
    const componentRepos = vertices.filter((v) =>
      component.nodes.find(
        (c) =>
          (c.type === NodeType.OUTPUT_REPO || c.type === NodeType.INPUT_REPO) &&
          c.name === v.name,
      ),
    );
    const id = jobSetId
      ? jobSetId
      : minBy(componentRepos, (r) => r.createdAt)?.name || '';

    const adjustedComponent = adjustDag(component, direction);

    return {
      id,
      nodes: adjustedComponent.nodes,
      links: adjustedComponent.links,
    };
  });
};

export default buildDags;
