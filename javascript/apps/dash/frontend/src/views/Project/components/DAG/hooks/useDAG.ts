import * as d3 from 'd3';
import {D3ZoomEvent} from 'd3';
import {useEffect, useState} from 'react';

import linkStateAsJobState from '@dash-frontend/lib/linkStateAsJobState';
import nodeStateAsPipelineState from '@dash-frontend/lib/nodeStateAsPipelineState';
import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import {
  Dag,
  JobState,
  Link,
  Node,
  NodeType,
  PipelineState,
} from '@graphqlTypes';

import busy from '../images/busy.svg';
import error from '../images/error.svg';
import noAccess from '../images/noAccess.svg';
import pipeline from '../images/pipeline.svg';
import repo from '../images/repo.svg';
import deriveRepoNameFromNode from '../utils/deriveRepoNameFromNode';

import useRouteController from './useRouteController';

const NODE_ICON_X_OFFSET = 20;
const NODE_ICON_Y_OFFSET = -10;
const NODE_IMAGE_Y_OFFSET = -20;
const NODE_IMAGE_PREVIEW_Y_OFFSET = -20;
const ORIGINAL_NODE_IMAGE_WIDTH = 170;
const ORIGINAL_NODE_IMAGE_HEIGHT = 102;
const NODE_TOOLTIP_WIDTH = 250;
const NODE_TOOLTIP_HEIGHT = 80;
const NODE_TOOLTIP_OFFSET = -88;

const convertNodeStateToDagState = (state: Node['state']) => {
  if (!state) return '';

  switch (nodeStateAsPipelineState(state)) {
    case PipelineState.PIPELINE_STANDBY:
    case PipelineState.PIPELINE_PAUSED:
      return 'idle';
    case PipelineState.PIPELINE_RUNNING:
    case PipelineState.PIPELINE_STARTING:
    case PipelineState.PIPELINE_RESTARTING:
      return 'busy';
    case PipelineState.PIPELINE_FAILURE:
    case PipelineState.PIPELINE_CRASHING:
      return 'error';
    default:
      return 'idle';
  }
};

const getLinkStyles = (d: Link) => {
  let className = 'link';
  if (linkStateAsJobState(d.state) === JobState.JOB_RUNNING)
    className = className.concat(' transferring');
  if (linkStateAsJobState(d.state) === JobState.JOB_FAILURE)
    className = className.concat(' error');

  className = className.concat(
    ` ${convertNodeStateToDagState(d.sourceState)}Source`,
  );
  className = className.concat(
    ` ${convertNodeStateToDagState(d.targetState)}Target`,
  );
  return className;
};

const getLineArray = (link: Link) => {
  const lineArray = link.bendPoints.reduce<[number, number][]>(
    (acc, point) => {
      acc.push([point.x, point.y]);
      return acc;
    },
    [[link.startPoint.x, link.startPoint.y]],
  );

  lineArray.push([link.endPoint.x, link.endPoint.y]);
  return lineArray;
};

const generateLinks = (
  svgParent: d3.Selection<SVGGElement, unknown, HTMLElement, unknown>,
  links: Link[],
) => {
  const link = svgParent
    .selectAll<SVGPathElement, Link>('.link')
    .data(links)
    .join<SVGPathElement>('path')
    .attr('d', (d) => d3.line()(getLineArray(d)))
    .attr('class', getLinkStyles)
    .attr('id', (d) => d.id)
    .attr('fill', 'none')
    .classed('link', true);

  // circle animates along path
  svgParent
    .selectAll<SVGCircleElement, Link>('.circle')
    .data(
      links.filter(
        (d) => linkStateAsJobState(d.state) === JobState.JOB_RUNNING,
      ),
    )
    .join<SVGCircleElement>('circle')
    .attr('r', 6)
    .attr('class', 'circle')
    .append<SVGAnimateMotionElement>('animateMotion')
    .attr('dur', '0.8s')
    .attr('repeatCount', 'indefinite')
    .append<SVGPathElement>('mpath')
    .attr('xlink:href', (d) => `#${d.id}`);

  return link;
};

const generateDefs = (
  svgParent: d3.Selection<SVGSVGElement, unknown, HTMLElement, unknown>,
) => {
  const defs = svgParent.append<SVGDefsElement>('svg:defs');
  const generateArrowBaseDef = (id: string, color: string) => {
    defs
      .append<SVGMarkerElement>('svg:marker')
      .attr('viewBox', '0 -5 10 10')
      .attr('refX', 9)
      .attr('markerWidth', 5)
      .attr('markerHeight', 5)
      .attr('orient', 'auto')
      .attr('id', id)
      .append<SVGPathElement>('svg:path')
      .attr('d', 'M0,-5L10,0L0,5')
      .attr('fill', color);
  };

  generateArrowBaseDef('end-arrow', '#000');
  generateArrowBaseDef('end-arrow-active', '#5ba3b1');
  generateArrowBaseDef('end-arrow-error', '#E02020');

  const generateSVGDef = (xml: XMLDocument, id: string) => {
    const SVG = d3
      .select<HTMLElement, unknown>(xml.documentElement)
      .attr('transform', null)
      .attr('class', id);
    defs
      .append(() => SVG.select<SVGGraphicsElement>('g').clone(true).node())
      .attr('id', id);
  };

  // load SVG images through d3.xml and set them as reusable defs
  Promise.all([
    d3.xml(repo),
    d3.xml(pipeline),
    d3.xml(noAccess),
    d3.xml(busy),
    d3.xml(error),
  ]).then((res) => {
    generateSVGDef(res[0], 'nodeImageRepo');
    generateSVGDef(res[1], 'nodeImagePipeline');
    generateSVGDef(res[2], 'nodeImageNoAccess');
    generateSVGDef(res[3], 'nodeIconBusy');
    generateSVGDef(res[4], 'nodeIconError');
  });
};

const generateNodeGroups = (
  svgParent: d3.Selection<SVGGElement, unknown, HTMLElement, unknown>,
  nodes: Node[],
  nodeWidth: number,
  nodeHeight: number,
  preview = false,
  handleSelectNode: (n: Node) => void,
  setHoveredNode: React.Dispatch<React.SetStateAction<string>>,
) => {
  const nodeGroup = svgParent
    .selectAll<SVGGElement, Node>('.nodeGroup')
    .data(nodes)
    .join((group) => {
      const enter = group
        .append<SVGGElement>('g')
        .attr('class', 'nodeGroup')
        .attr('id', (d) => `${d.name}GROUP`)
        .attr('transform', (d) => `translate (${d.x}, ${d.y})`)
        .on('click', (_event, d) => handleSelectNode(d))
        .on('mouseover', (_event, d) => setHoveredNode(d.name))
        .on('mouseout', () => setHoveredNode(''));

      !preview &&
        enter
          .append('foreignObject')
          .attr('class', 'node')
          .attr('width', nodeWidth)
          .attr('height', nodeHeight)
          .append('xhtml:span')
          .html(
            (d) =>
              `<p class="label" style="height:${nodeHeight}px">${
                d.access ? deriveRepoNameFromNode(d) : 'No Access'
              }</p>`,
          );

      !preview &&
        enter
          .append('foreignObject')
          .attr('class', 'nodeTooltip')
          .attr('width', NODE_TOOLTIP_WIDTH)
          .attr('height', NODE_TOOLTIP_HEIGHT)
          .style('min-height', NODE_TOOLTIP_HEIGHT)
          .attr('y', NODE_TOOLTIP_OFFSET)
          .attr('x', -(NODE_TOOLTIP_WIDTH / 4))
          .append('xhtml:span')
          .html(
            (d) => `<p class="tooltipText" style="min-height: ${
              NODE_TOOLTIP_HEIGHT - 9
            }px">
                ${deriveRepoNameFromNode(d)}
                <br /><br />
                ${
                  d.type === NodeType.Pipeline
                    ? `${d.type.toLowerCase()} status: ${readablePipelineState(
                        d.state || '',
                      )}`
                    : ''
                }
              </p>`,
          );

      enter
        .append<SVGUseElement>('use')
        .attr('xlink:href', (d) => {
          if (d.access) {
            if (d.type === NodeType.Repo) return '#nodeImageRepo';
            if (d.type === NodeType.Pipeline) return '#nodeImagePipeline';
          }
          return '#nodeImageNoAccess';
        })
        .attr('transform', `scale(${nodeHeight / ORIGINAL_NODE_IMAGE_HEIGHT})`)
        .attr('x', () => (ORIGINAL_NODE_IMAGE_WIDTH - nodeWidth) / 2)
        .attr('y', (d) =>
          !preview ? NODE_IMAGE_Y_OFFSET : NODE_IMAGE_PREVIEW_Y_OFFSET,
        )
        .attr('pointer-events', 'none');

      enter
        .append<SVGUseElement>('use')
        .attr('xlink:href', (d) => {
          const state = convertNodeStateToDagState(d.state);
          if (state === 'busy') return '#nodeIconBusy';
          if (state === 'error') return '#nodeIconError';
          return null;
        })
        .attr('x', nodeWidth - NODE_ICON_X_OFFSET)
        .attr('y', NODE_ICON_Y_OFFSET)
        .attr('pointer-events', 'none');
      return enter;
    });

  return nodeGroup;
};

type useDAGProps = {
  id: string;
  svgParentSize: {width: number; height: number};
  setSVGParentSize: React.Dispatch<
    React.SetStateAction<{
      width: number;
      height: number;
    }>
  >;
  nodeWidth: number;
  nodeHeight: number;
  data: Dag;
  preview: boolean;
};

const useDAG = ({
  id,
  data,
  svgParentSize,
  setSVGParentSize,
  nodeWidth,
  nodeHeight,
  preview = false,
}: useDAGProps) => {
  const {selectedNode, navigateToNode} = useRouteController({dag: data});
  const [hoveredNode, setHoveredNode] = useState('');

  // Pre-build steps
  useEffect(() => {
    const svg = d3.select<SVGSVGElement, unknown>(`#${id}`);

    // Define reusable SVG elements, i.e. different arrow heads
    generateDefs(svg);
  }, [id]);

  // build dag
  useEffect(() => {
    const svg = d3.select<SVGSVGElement, unknown>(`#${id}`);
    const graph = d3.select<SVGGElement, unknown>(`#${id}Graph`);

    const links = generateLinks(graph, data.links);
    generateNodeGroups(
      graph,
      data.nodes,
      nodeWidth,
      nodeHeight,
      preview,
      navigateToNode,
      setHoveredNode,
    );

    const zoomed = (event: D3ZoomEvent<SVGSVGElement, unknown>) => {
      const {transform} = event;
      graph.attr('transform', transform.toString());
    };

    const zoom = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.1, 4])
      .on('zoom', zoomed);

    svg.call(zoom);

    links.attr('class', (d) => `${getLinkStyles(d)}`);

    // initialize zoom based on node positions and center dag
    const xExtent = d3.extent(data.nodes, (d) => d.x);
    const yExtent = d3.extent(data.nodes, (d) => d.y);
    const xMin = xExtent[0] || 0;
    const xMax = xExtent[1] || svgParentSize.width;
    const yMin = yExtent[0] || 0;
    const yMax = yExtent[1] || svgParentSize.height;

    const transform = d3.zoomIdentity
      .translate(
        svgParentSize.width / 2,
        svgParentSize.height / 2 - nodeHeight / 2,
      )
      .translate(-(xMin + xMax) / 2, -(yMin + yMax) / 2);
    svg.call(zoom.transform, transform);
  }, [id, nodeHeight, nodeWidth, data, svgParentSize, preview, navigateToNode]);

  // Update node classes based on react state
  useEffect(() => {
    const graph = d3.select<SVGGElement, unknown>(`#${id}Graph`);

    !preview &&
      graph
        .selectAll<SVGGElement, Node>('.nodeGroup')
        .attr(
          'class',
          (d) =>
            `${`nodeGroup ${convertNodeStateToDagState(d.state)}`} ${
              [selectedNode?.name, hoveredNode].includes(d.name)
                ? 'selected'
                : ''
            }`,
        );

    !preview &&
      graph.selectAll<SVGPathElement, Link>('.link').attr('class', (d) => {
        return `${getLinkStyles(d)} ${
          [selectedNode?.name, hoveredNode].includes(d.source) ||
          [selectedNode?.name, hoveredNode].includes(d.target)
            ? 'selected'
            : ''
        }`;
      });
  }, [id, preview, selectedNode, hoveredNode]);

  // Post-build steps
  useEffect(() => {
    const svgElement = d3.select<SVGSVGElement, null>(`#${id}`).node();

    // Set parent svg with and height to our SVG's content-defined bounding box to readjust graph positioning
    if (svgElement) {
      setSVGParentSize({
        width: svgElement.getBBox().width + nodeWidth,
        height: svgElement.getBBox().width + nodeHeight,
      });
    }
  }, [id, setSVGParentSize, nodeHeight, nodeWidth]);
};

export default useDAG;
