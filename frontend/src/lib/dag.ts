/* eslint-disable import/no-named-as-default */
import {NodeType, Vertex, NodeState} from '@graphqlTypes';
import ELK from 'elkjs/lib/elk.bundled.js';
import keyBy from 'lodash/keyBy';
import objectHash from 'object-hash';

import {NODE_INPUT_REPO} from '@dash-frontend/views/Project/constants/nodeSizes';

import {LinkInputData, NodeInputData, Link, Node, DagDirection} from './types';

const normalizeData = async (
  vertices: Vertex[],
  nodeWidth: number,
  nodeHeight: number,
  direction: DagDirection = DagDirection.DOWN,
) => {
  const elk = new ELK();

  // Calculate the indicies of the nodes in the DAG, as
  // the "link" object requires the source & target attributes
  // to reference the index from the "nodes" list. This prevents
  // us from needing to run "indexOf" on each individual node
  // when building the list of links for the client.
  const nodeIndexMap: Record<string, number> = {};
  for (let i = 0; i < vertices.length; i++) {
    nodeIndexMap[vertices[i].id] = i;
  }

  const vertMap = keyBy(vertices, (v) => v.id);

  // create elk edges
  const elkEdges: LinkInputData[] = [];
  for (const vertex of vertices) {
    const parentStrings = vertex.parents.map(({id}) => {
      return id;
    });
    const uniqueParents = new Set(parentStrings);
    uniqueParents.forEach((vertexParentName) => {
      const parentIndex = nodeIndexMap[vertexParentName];
      const parentVertex = vertices[parentIndex];

      // edge case: if pipeline has input repos that cannot be found, the link is not valid
      // it might be in another project
      if (
        vertex.type === NodeType.PIPELINE &&
        !(vertexParentName in nodeIndexMap)
      ) {
        return;
      }

      elkEdges.push({
        // hashing here for consistency
        id: objectHash({node: vertex, vertexParentName}),
        sources: [vertexParentName],
        targets: [vertex.id],
        state: parentVertex?.jobNodeState || undefined,
        sourceState: parentVertex?.nodeState || undefined,
        targetState: vertex.nodeState || undefined,
        sections: [],
        transferring: vertex.jobNodeState === NodeState.RUNNING,
      });
    });
  }

  // create elk children
  const elkChildren = vertices.map<NodeInputData>((node) => ({
    id: node.id,
    project: node.project,
    name: node.name,
    type: node.type,
    state: node.state || undefined,
    jobState: node.jobState || undefined,
    nodeState: node.nodeState || undefined,
    jobNodeState: node.jobNodeState || undefined,
    access: node.access,
    x: 0,
    y: 0,
    width: nodeWidth,
    height: node.type === NodeType.REPO ? NODE_INPUT_REPO : nodeHeight,
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
        'org.eclipse.elk.aspectRatio': horizontal ? '0.2' : '10',
        'org.eclipse.elk.mergeEdges': 'false',
        'org.eclipse.elk.direction': direction,
        'org.eclipse.elk.layered.layering.strategy': 'INTERACTIVE',
        'spacing.componentComponent': horizontal ? '75' : '150',

        'org.eclipse.elk.layered.spacing.edgeNodeBetweenLayers': '20',
        'org.eclipse.elk.layered.spacing.nodeNodeBetweenLayers': '40',
      },
    },
  );

  // convert elk edges to graphql Link
  const links = elkEdges.map<Link>(
    ({
      sections,
      sources,
      targets,
      junctionPoints: _junctionPoints,
      labels: _labels,
      layoutOptions: _layoutOptions,
      ...rest
    }) => {
      const sourceId = sources[0];
      const targetId = targets[0];
      const selection = sections?.[0] || {
        startPoint: {x: 0, y: 0},
        endPoint: {x: 0, y: 0},
        bendPoints: [],
      };

      return {
        ...rest,
        source: {id: sourceId, name: vertMap[sourceId].name},
        target: {id: targetId, name: vertMap[targetId].name},
        startPoint: selection.startPoint,
        endPoint: selection.endPoint,
        bendPoints: selection.bendPoints || [],
        isCrossProject: vertMap[sourceId].type === NodeType.CROSS_PROJECT_REPO,
      };
    },
  );

  // convert elk children to graphql Node
  const nodes = elkChildren.map<Node>((node) => ({
    id: node.id,
    project: node.project,
    name: node.name,
    x: node.x || 0,
    y: node.y || 0,
    type: node.type,
    state: node.state,
    jobState: node.jobState,
    nodeState: node.nodeState,
    jobNodeState: node.jobNodeState,
    access: node.access,
  }));

  return {
    nodes,
    links,
  };
};

const buildDags = async (
  vertices: Vertex[],
  nodeWidth: number,
  nodeHeight: number,
  direction = DagDirection.DOWN,
  setDagError: React.Dispatch<React.SetStateAction<string | undefined>>,
) => {
  try {
    return await normalizeData(vertices, nodeWidth, nodeHeight, direction);
  } catch (e) {
    console.error(e);
    setDagError(`Unable to construct DAG from repos and pipelines.`);
  }
};

export default buildDags;
