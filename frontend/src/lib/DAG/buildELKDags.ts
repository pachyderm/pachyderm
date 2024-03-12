/* eslint-disable import/no-named-as-default */
import {Graph} from '@dagrejs/graphlib';
import ELK from 'elkjs/lib/elk.bundled.js';
import keyBy from 'lodash/keyBy';
import objectHash from 'object-hash';

import {
  LinkInputData,
  NodeInputData,
  Link,
  Node,
  DagDirection,
  NodeState,
  NodeType,
} from '@dash-frontend/lib/types';
import {
  CONNECTED_EGRESS_NODE_HEIGHT,
  NODE_HEIGHT,
  NODE_INPUT_REPO,
  NODE_WIDTH,
} from '@dash-frontend/views/Project/constants/nodeSizes';

import {Vertex} from './convertPipelinesAndReposToVertices';

const getNodeHeight = (type: NodeType) => {
  switch (type) {
    case NodeType.REPO:
    case NodeType.CRON_REPO:
      return NODE_INPUT_REPO;
    case NodeType.CROSS_PROJECT_REPO:
    case NodeType.EGRESS:
      return CONNECTED_EGRESS_NODE_HEIGHT;
    default:
      return NODE_HEIGHT;
  }
};

const normalizeData = async (
  vertices: Vertex[],
  direction: DagDirection = DagDirection.DOWN,
) => {
  const elk = new ELK();
  // generate @dagrejs/graphlib for highlighting links based on sub-dags
  const graph = new Graph();
  const reverseGraph = new Graph();

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
    graph.setNode(vertex.id, vertex);
    reverseGraph.setNode(vertex.id, vertex);
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
      graph.setEdge(vertexParentName, vertex.id);
      reverseGraph.setEdge(vertex.id, vertexParentName);
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
    hasCronInput: node.hasCronInput,
    parallelism: node.parallelism,
    egressType: node.egressType,
    x: 0,
    y: 0,
    width: NODE_WIDTH,
    height: getNodeHeight(node.type),
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
        startPoint: {
          x: Math.round(selection.startPoint.x),
          y: Math.round(selection.startPoint.y),
        },
        endPoint: {
          x: Math.round(selection.endPoint.x),
          y: Math.round(selection.endPoint.y),
        },
        bendPoints: selection.bendPoints
          ? selection.bendPoints?.map((point) => {
              return {x: Math.round(point.x), y: Math.round(point.y)};
            })
          : [],
        isCrossProject: vertMap[sourceId].type === NodeType.CROSS_PROJECT_REPO,
      };
    },
  );

  // convert elk children to graphql Node
  const nodes = elkChildren.map<Node>((node) => ({
    id: node.id,
    project: node.project,
    name: node.name,
    x: Math.round(node.x || 0),
    y: Math.round(node.y || 0),
    type: node.type,
    state: node.state,
    jobState: node.jobState,
    nodeState: node.nodeState,
    jobNodeState: node.jobNodeState,
    access: node.access,
    hasCronInput: node.hasCronInput,
    parallelism: node.parallelism,
    egressType: node.egressType,
  }));
  return {
    graph,
    reverseGraph,
    nodes,
    links,
  };
};

const buildELKDags = async (
  vertices: Vertex[],
  direction = DagDirection.DOWN,
  setDagError: React.Dispatch<React.SetStateAction<string | undefined>>,
) => {
  try {
    return await normalizeData(vertices, direction);
  } catch (e) {
    console.error(e);
    setDagError(`Unable to construct DAG from repos and pipelines.`);
  }
};

export default buildELKDags;
