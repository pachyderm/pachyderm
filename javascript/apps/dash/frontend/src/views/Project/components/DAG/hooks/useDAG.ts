import * as d3 from 'd3';
import {D3ZoomEvent} from 'd3';
import {useEffect} from 'react';

import {LinkDatum, NodeDatum} from '@dash-frontend/lib/DAGTypes';
import linkStateAsJobState from '@dash-frontend/lib/linkStateAsJobState';
import nodeStateAsPipelineState from '@dash-frontend/lib/nodeStateAsPipelineState';
import {Dag, JobState, NodeType, PipelineState} from '@graphqlTypes';

import checkmark from '../images/checkmark.svg';
import error from '../images/error.svg';
import noAccess from '../images/noAccess.svg';
import pipeline from '../images/pipeline.svg';
import repo from '../images/repo.svg';

const getLinkStyles = (d: LinkDatum) => {
  let className = 'link';
  if (linkStateAsJobState(d.state) === JobState.JOB_RUNNING)
    className = className.concat(' transferring');
  if (linkStateAsJobState(d.state) === JobState.JOB_FAILURE)
    className = className.concat(' error');
  return className;
};

const transformArrow = (d: LinkDatum, nodeRadius: number, invert?: boolean) => {
  // This function returns the x and y translation needed to move the end
  // point of a link to the edge of a target square. A trigonometric function
  // is used to determine the length of the hypotenuse of a right triangle formed by
  // the angle of the link line and the square width.

  // Type assertion to NodeDatum for the link source and target are used here
  // because during initialization the source and target properties for
  // SimulationNodeDatum can be a string or number but once we apply the force
  // d3 reassigns those values with a reference to the actual node so the type
  // is string | number | NodeDatum.
  // The type for SimulationNodeDatum can be found here:
  // https://github.com/DefinitelyTyped/DefinitelyTyped/blob/master/types/d3-force/index.d.ts#L72
  const end = invert ? (d.source as NodeDatum) : (d.target as NodeDatum);
  const start = invert ? (d.target as NodeDatum) : (d.source as NodeDatum);
  const squareAngle = Math.PI / 2;

  const dx = (start.x || 0) - (end.x || 0);
  const dy = (start.y || 0) - (end.y || 0);
  const rAngle = Math.abs(Math.atan2(-dx, -dy) % squareAngle);
  const inverseAngle = squareAngle - (rAngle % squareAngle);

  const offset =
    (Math.sin(squareAngle) * nodeRadius) /
    Math.sin(squareAngle - Math.min(inverseAngle, rAngle));
  const scale = offset / Math.sqrt(dx * dx + dy * dy);
  const xTranslation = (end.x || 0) + dx * scale;
  const yTranslation = (end.y || 0) + dy * scale;

  return {
    x: xTranslation,
    y: yTranslation,
  };
};

const generateLinks = (
  svgParent: d3.Selection<SVGGElement, unknown, HTMLElement, unknown>,
  links: LinkDatum[],
) => {
  const link = svgParent
    .selectAll<SVGPathElement, LinkDatum>('.link')
    .data(links)
    .join<SVGPathElement>('path')
    .attr('class', getLinkStyles)
    .attr('id', (d) => `${d.source}-${d.target}`)
    .classed('link', true);

  // circle animates along path
  svgParent
    .selectAll<SVGCircleElement, LinkDatum>('.circle')
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
    .attr('xlink:href', (d) => `#${d.source}-${d.target}`);

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
      .attr('refX', 6)
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
    d3.xml(checkmark),
    d3.xml(error),
  ]).then((res) => {
    generateSVGDef(res[0], 'nodeImageRepo');
    generateSVGDef(res[1], 'nodeImagePipeline');
    generateSVGDef(res[2], 'nodeImageNoAccess');
    generateSVGDef(res[3], 'nodeIconSuccess');
    generateSVGDef(res[4], 'nodeIconError');
  });
};

const generateNodeGroups = (
  svgParent: d3.Selection<SVGGElement, unknown, HTMLElement, unknown>,
  nodes: NodeDatum[],
  nodeWidth: number,
  nodeHeight: number,
  preview = false,
) => {
  const nodeGroup = svgParent
    .selectAll<SVGGElement, NodeDatum>('.nodeGroup')
    .data(nodes)
    .join((group) => {
      const enter = group
        .append<SVGGElement>('g')
        .attr('class', `nodeGroup ${!preview ? 'draggable' : ''}`)
        .attr('id', (group) => `${group.name}GROUP`);

      enter
        .append<SVGRectElement>('rect')
        .attr('class', 'node')
        .attr('id', (d) => d.name)
        .attr('width', () => nodeWidth * 0.95)
        .attr('height', () => nodeHeight * 0.95)
        .attr('y', nodeHeight * 0.2);

      !preview &&
        enter
          .append<SVGTextElement>('text')
          .attr('class', 'label')
          .attr('text-anchor', 'middle')
          .attr('fill', (d) => {
            const state = nodeStateAsPipelineState(d.state);
            if (
              d.access &&
              (state === PipelineState.PIPELINE_CRASHING ||
                state === PipelineState.PIPELINE_FAILURE)
            )
              return '#E02020';
            return '#020408';
          })
          .text((d) => (d.access ? d.name : 'No Access'))
          .attr('y', nodeHeight * 0.2 + (nodeHeight * 0.95) / 2 + 15)
          .attr('x', nodeWidth / 2);

      enter
        .append<SVGUseElement>('use')
        .attr('xlink:href', (d) => {
          if (d.access) {
            if (d.type === NodeType.Repo) return '#nodeImageRepo';
            if (d.type === NodeType.Pipeline) return '#nodeImagePipeline';
          }
          return '#nodeImageNoAccess';
        })
        .attr('transform', `scale(${nodeHeight / 102})`)
        .attr('x', () => (170 - nodeWidth) / 2)
        .attr('y', (d) => (!preview ? (d.access ? -4 : 6) : nodeHeight - 10));

      enter
        .append<SVGUseElement>('use')
        .attr('xlink:href', (d) => {
          if (d.access) {
            const state = nodeStateAsPipelineState(d.state);
            if (state === PipelineState.PIPELINE_RUNNING)
              // is running considered success?
              return '#nodeIconSuccess';
            if (
              state === PipelineState.PIPELINE_CRASHING ||
              state === PipelineState.PIPELINE_FAILURE
            )
              return '#nodeIconError';
          }
          return null;
        })
        .attr('x', nodeWidth * 0.95 - 15)
        .attr('y', nodeHeight - nodeHeight * 0.95);
      return enter;
    });

  return nodeGroup;
};

const assignPositions = (
  svgParent: d3.Selection<SVGGElement, unknown, HTMLElement, unknown>,
  links: d3.Selection<SVGPathElement, LinkDatum, SVGGElement, unknown>,
  nodes: NodeDatum[],
  nodeWidth: number,
  nodeHeight: number,
) => {
  links.attr('d', (d) => {
    const start = transformArrow(d, nodeWidth / 2, true);
    const end = transformArrow(d, nodeWidth / 2);
    return `M ${start.x} ${start.y} L ${end.x} ${end.y}`;
  });
  const nodeGroups = svgParent
    .selectAll<SVGGraphicsElement, NodeDatum>('.nodeGroup')
    .data(nodes);
  nodeGroups.attr('transform', (d: NodeDatum) => {
    return `translate(${d.x ? d.x - (nodeWidth / 2) * 0.95 : 0}, ${
      d.y ? d.y - ((nodeHeight * 1.2) / 2) * 0.95 : 0
    })`;
  });
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
  // Pre-build steps
  useEffect(() => {
    const svg = d3.select<SVGSVGElement, unknown>(`#${id}`);

    // Define reusable SVG elements, i.e. different arrow heads
    generateDefs(svg);
  }, [id]);

  // Build d3-force graph
  useEffect(() => {
    const defaultd3Node: d3.SimulationNodeDatum = {
      fx: undefined,
      fy: undefined,
      index: undefined,
      vx: undefined,
      vy: undefined,
      x: undefined,
      y: undefined,
    };
    const svg = d3.select<SVGSVGElement, unknown>(`#${id}`);
    const graph = d3.select<SVGGElement, unknown>(`#${id}Graph`);
    const d3Links: LinkDatum[] = data.links.map((link) => ({
      ...link,
      index: undefined,
    }));
    const d3Nodes: NodeDatum[] = data.nodes.map((node) => ({
      ...node,
      ...defaultd3Node,
    }));
    const links = generateLinks(graph, d3Links);
    generateNodeGroups(graph, d3Nodes, nodeWidth, nodeHeight, preview);

    const simulation = d3
      .forceSimulation<NodeDatum, LinkDatum>()
      .nodes(d3Nodes)
      .force('collide', d3.forceCollide(nodeWidth * 1.1))
      .force('centerX', d3.forceX(svgParentSize.width / 2))
      .force('centerY', d3.forceY(svgParentSize.height / 2))
      .force('link', d3.forceLink(d3Links))
      .on('tick', () => {
        assignPositions(graph, links, d3Nodes, nodeWidth, nodeHeight);
      })
      .stop();

    simulation.tick(300);

    const enableDragging = (
      simulation: d3.Simulation<NodeDatum, LinkDatum>,
    ) => {
      function dragged(
        event: d3.D3DragEvent<SVGElement, NodeDatum, NodeDatum>,
        d: NodeDatum,
      ) {
        const clamp = (x: number, low: number, high: number) => {
          if (x < low) return low;
          if (x > high) return high;
          return x;
        };

        const graphNode = graph.node();
        if (graphNode) {
          // Grab the transform on the graph group and use that to invert the drag edges.
          const graphTransform = d3.zoomTransform(graphNode);
          const left = graphTransform.invertX(
            (nodeWidth / 2) * graphTransform.k,
          );
          const right = graphTransform.invertX(
            svgParentSize.width - (nodeWidth / 2) * graphTransform.k,
          );
          const top = graphTransform.invertY(
            nodeHeight * graphTransform.k + 12,
          );
          const bottom = graphTransform.invertY(
            svgParentSize.height - (nodeHeight / 2 + 10) * graphTransform.k,
          );
          d.fx = clamp(event.x, left, right);
          d.fy = clamp(event.y, top, bottom);
          simulation.alpha(0).force('collide', d3.forceCollide()).restart();
        }
      }
      const drag = d3.drag<SVGElement, NodeDatum>().on('drag', dragged);
      svg
        .selectAll<SVGElement, NodeDatum>('.nodeGroup')
        .data(d3Nodes)
        .call(drag);
    };

    const zoomed = (event: D3ZoomEvent<SVGSVGElement, unknown>) => {
      const {transform} = event;
      graph.attr('transform', transform.toString());
    };

    const zoom = d3
      .zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.1, 4])
      .on('zoom', zoomed);

    svg.call(zoom);

    !preview && enableDragging(simulation);
    assignPositions(graph, links, d3Nodes, nodeWidth, nodeHeight);

    // initialize zoom based on node positions and center dag
    const xExtent = d3.extent(d3Nodes, (d) => d.x);
    const yExtent = d3.extent(d3Nodes, (d) => d.y);
    const xMin = xExtent[0] || 0;
    const xMax = xExtent[1] || svgParentSize.width;
    const yMin = yExtent[0] || 0;
    const yMax = yExtent[1] || svgParentSize.height;
    const xScale = svgParentSize.width / (xMax - xMin);
    const yScale = svgParentSize.height / 2 / (yMax - yMin);
    const minScale = Math.min(xScale, yScale);

    const transform = d3.zoomIdentity
      .translate(svgParentSize.width / 2, svgParentSize.height / 2)
      .scale(minScale)
      .translate(-(xMin + xMax) / 2, -(yMin + yMax) / 2);
    svg.call(zoom.transform, transform);
  }, [id, nodeHeight, nodeWidth, data, svgParentSize, preview]);

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
