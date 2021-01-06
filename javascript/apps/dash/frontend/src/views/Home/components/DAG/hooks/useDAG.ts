/* eslint-disable @typescript-eslint/no-explicit-any */
import * as d3 from 'd3';
import {useEffect} from 'react';

import {
  LinkDatum,
  LinkInputData,
  NodeDatum,
  DataProps,
  Coordinate,
} from 'lib/DAGTypes';

import checkmark from '../images/checkmark.svg';
import error from '../images/error.svg';
import noAccess from '../images/noAccess.svg';
import pipeline from '../images/pipeline.svg';
import repo from '../images/repo.svg';

const getLinkStyles = (d: LinkInputData) => {
  let className = 'link';
  if (d.active) className = className.concat(' transferring');
  if (d.error) className = className.concat(' error');
  return className;
};

const transformArrow = (d: LinkDatum, nodeRadius: number, invert?: boolean) => {
  // This function returns the x and y translation needed to move the end
  // point of a link to the edge of a target square. A trigonometric function
  // is used to determine the length of the hypotenuse of a right triangle formed by
  // the angle of the link line and the square width.
  const end = invert ? d.source : d.target;
  const start = invert ? d.target : d.source;
  const squareAngle = Math.PI / 2;

  const dx = start.x - end.x;
  const dy = start.y - end.y;
  const rAngle = Math.abs(Math.atan2(-dx, -dy) % squareAngle);
  const inverseAngle = squareAngle - (rAngle % squareAngle);

  const offset =
    (Math.sin(squareAngle) * nodeRadius) /
    Math.sin(squareAngle - Math.min(inverseAngle, rAngle));
  const scale = offset / Math.sqrt(dx * dx + dy * dy);
  const xTranslation = end.x + dx * scale;
  const yTranslation = end.y + dy * scale;

  return {
    x: xTranslation,
    y: yTranslation,
  };
};

const generateLinks = (
  svgParent: d3.Selection<d3.BaseType, unknown, HTMLElement, any>,
  links: LinkInputData[],
) => {
  const link = svgParent
    .selectAll('.link')
    .data(links)
    .join('path')
    .attr('class', getLinkStyles)
    .attr('id', (d: LinkInputData) => `${d.source}-${d.target}`)
    .classed('link', true);

  // circle animates along path
  svgParent
    .selectAll('.circle')
    .data(links.filter((d) => d.active))
    .join('circle')
    .attr('r', 6)
    .attr('class', 'circle')
    .append('animateMotion')
    .attr('dur', '0.8s')
    .attr('repeatCount', 'indefinite')
    .append('mpath')
    .attr('xlink:href', (d: LinkInputData) => `#${d.source}-${d.target}`);

  return link;
};

const generateDefs = (
  svgParent: d3.Selection<d3.BaseType, unknown, HTMLElement, any>,
) => {
  const defs = svgParent.append('svg:defs');
  const generateArrowBaseDef = (id: string, color: string) => {
    defs
      .append('svg:marker')
      .attr('viewBox', '0 -5 10 10')
      .attr('refX', 6)
      .attr('markerWidth', 5)
      .attr('markerHeight', 5)
      .attr('orient', 'auto')
      .attr('id', id)
      .append('svg:path')
      .attr('d', 'M0,-5L10,0L0,5')
      .attr('fill', color);
  };

  generateArrowBaseDef('end-arrow', '#000');
  generateArrowBaseDef('end-arrow-active', '#5ba3b1');
  generateArrowBaseDef('end-arrow-error', '#E02020');

  const generateSVGDef = (xml: XMLDocument, id: string) => {
    const SVG = d3
      .select(xml.documentElement)
      .attr('transform', null)
      .attr('class', id);
    defs.append(() => SVG.select('g').clone(true).node()).attr('id', id);
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
  svgParent: d3.Selection<d3.BaseType, unknown, HTMLElement, any>,
  nodes: NodeDatum[],
  nodeWidth: number,
  nodeHeight: number,
) => {
  const nodeGroup = svgParent
    .selectAll('.nodeGroup')
    .data(nodes)
    .join((group) => {
      const enter = group.append('g').attr('class', 'nodeGroup');

      enter
        .append('rect')
        .attr('class', 'node')
        .attr('id', (d: NodeDatum) => d.name)
        .attr('width', () => nodeWidth * 0.95)
        .attr('height', () => nodeHeight * 0.95)
        .attr('y', nodeHeight * 0.2);

      enter
        .append('text')
        .attr('class', 'label')
        .attr('text-anchor', 'middle')
        .attr('fill', (d: NodeDatum) =>
          d.access && d.error ? '#E02020' : '#020408',
        )
        .text((d: NodeDatum) => (d.access ? d.name : 'No Access'))
        .attr('y', nodeHeight * 0.2 + (nodeHeight * 0.95) / 2 + 15)
        .attr('x', nodeWidth / 2);

      enter
        .append('use')
        .attr('xlink:href', (d) => {
          if (d.access) {
            if (d.type === 'repo') return '#nodeImageRepo';
            if (d.type === 'pipeline') return '#nodeImagePipeline';
          }
          return '#nodeImageNoAccess';
        })
        .attr('y', (d) => (d.access ? -4 : 6));

      enter
        .append('use')
        .attr('xlink:href', (d) => {
          if (d.access) {
            if (d.success) return '#nodeIconSuccess';
            if (d.error) return '#nodeIconError';
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
  svgParent: d3.Selection<d3.BaseType, unknown, HTMLElement, any>,
  links: d3.Selection<
    Element | d3.EnterElement | Document | Window | SVGPathElement | null,
    LinkInputData,
    d3.BaseType,
    unknown
  >,
  nodes: NodeDatum[],
  nodeWidth: number,
  nodeHeight: number,
) => {
  links.attr('d', (d: any) => {
    const start = transformArrow(d, nodeWidth / 2, true);
    const end = transformArrow(d, nodeWidth / 2);
    return `M ${start.x} ${start.y} L ${end.x} ${end.y}`;
  });
  const nodeGroups = svgParent.selectAll('.nodeGroup').data(nodes);
  nodeGroups.attr('transform', (d: NodeDatum) => {
    return `translate(${d.x ? d.x - (nodeWidth / 2) * 0.95 : 0}, ${
      d.y ? d.y - ((nodeHeight * 1.2) / 2) * 0.95 : 0
    })`;
  });
};

const attachBackgroundDragHandlers = (baseElement: HTMLElement) => {
  // https://htmldom.dev/drag-to-scroll/
  let positionState = {top: 0, left: 0, x: 0, y: 0};
  const mouseDownHandler = (e: MouseEvent) => {
    positionState = {
      left: baseElement.scrollLeft,
      top: baseElement.scrollTop,
      // Get the current mouse position
      x: e.clientX,
      y: e.clientY,
    };

    document.addEventListener('mousemove', mouseMoveHandler);
    document.addEventListener('mouseup', mouseUpHandler);
  };

  const mouseMoveHandler = (e: MouseEvent) => {
    // How far the mouse has been moved
    const dx = e.clientX - positionState.x;
    const dy = e.clientY - positionState.y;

    // Scroll the baseElementment
    baseElement.scrollTop = positionState.top - dy;
    baseElement.scrollLeft = positionState.left - dx;
  };

  const mouseUpHandler = () => {
    document.removeEventListener('mousemove', mouseMoveHandler);
    document.removeEventListener('mouseup', mouseUpHandler);
  };

  baseElement.addEventListener('mousedown', mouseDownHandler);
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
  data: DataProps;
};

const useDAG = ({
  id,
  data,
  svgParentSize,
  setSVGParentSize,
  nodeWidth,
  nodeHeight,
}: useDAGProps) => {
  // Pre-build steps
  useEffect(() => {
    const svg = d3.select(`#${id}`);

    // Define reusable SVG elements, i.e. different arrow heads
    generateDefs(svg);
  }, [id]);

  // Build d3-force graph
  useEffect(() => {
    const svg = d3.select(`#${id}`);
    const links = generateLinks(svg, data.links);
    generateNodeGroups(svg, data.nodes, nodeWidth, nodeHeight);

    const simulation = d3
      .forceSimulation()
      .nodes(data.nodes)
      .force('collide', d3.forceCollide(nodeWidth * 1.1))
      .force('centerX', d3.forceX(svgParentSize.width / 2))
      .force('centerY', d3.forceY(svgParentSize.height / 2))
      .force('link', d3.forceLink(data.links))
      .on('tick', () => {
        assignPositions(svg, links, data.nodes, nodeWidth, nodeHeight);
      })
      .stop()
      .tick(300);

    const enableDragging = (simulation: any) => {
      const clamp = (x: number, lo: number, hi: number) =>
        x < lo ? lo : x > hi ? hi : x;

      function dragged(event: Coordinate, d: NodeDatum) {
        d.fx = clamp(event.x, 0, svgParentSize.width);
        d.fy = clamp(event.y, 0, svgParentSize.height);
        simulation.alpha(0).force('collide', d3.forceCollide()).restart();
      }
      const drag = d3.drag<any, NodeDatum, d3.BaseType>().on('drag', dragged);
      svg.selectAll('.nodeGroup').data(data.nodes).call(drag);
    };
    enableDragging(simulation);
    assignPositions(svg, links, data.nodes, nodeWidth, nodeHeight);
  }, [id, nodeHeight, nodeWidth, data, svgParentSize]);

  // Post-build steps
  useEffect(() => {
    const svgElement = document.getElementById(id) as SVGGraphicsElement | null;
    const baseElement = document.getElementById(`${id}_base`);

    // Set parent svg with and height to our SVG's content-defined bounding box to readjust graph positioning
    svgElement &&
      setSVGParentSize({
        width: svgElement.getBBox().width + 600,
        height: svgElement.getBBox().height + 300,
      });
    // Attach drag handlers to move overall SVG around parent container for browser only, NOT individual nodes
    baseElement && attachBackgroundDragHandlers(baseElement);
  }, [id, setSVGParentSize]);
};

export default useDAG;
