import {format} from 'date-fns';

import MontserratFont from './assets/MontserratFont_Base64';
import PublicFont from './assets/PublicFont_Base64';

const LABEL_PADDING = 60;
const CONTAINER_ELEMENTS = ['svg', 'g', 'foreignObject', 'SPAN'];
const RELEVANT_STYLES: Record<string, string[]> = {
  rect: [
    'fill',
    'stroke',
    'stroke-width',
    'stroke-dasharray',
    'border-radius',
    'paint-order',
    'filter',
  ],
  use: [
    'fill',
    'stroke',
    'stroke-width',
    'border-radius',
    'paint-order',
    'filter',
    'marker-end',
  ],
  path: ['fill', 'stroke', 'stroke-width', 'marker-end', 'stroke-linejoin'],
  circle: ['fill', 'stroke', 'stroke-width'],
  line: ['stroke', 'stroke-width'],
  text: [
    'fill',
    'font-size',
    'text-anchor',
    'font-family',
    'font-weight',
    'line-height',
    'letter-spacing',
  ],
  polygon: ['stroke', 'fill'],
  P: ['opacity'],
};

const addLabel = (node: Node, text: string, line = 1) => {
  const newText = document.createElementNS(
    'http://www.w3.org/2000/svg',
    'text',
  );
  newText.setAttribute('x', '20');
  newText.setAttribute('y', `${25 * line - 15}`);
  newText.setAttribute('fill', '#747475');
  newText.setAttribute('font-family', 'Public Sans');
  newText.setAttribute('font-size', '14');
  newText.textContent = text;
  node.appendChild(newText);
};

// we need to parse for each computed style and add them manually to SVG children outside of our regular React/Css Modules
const computeStyles = (ParentNode: Node, OrigNode: Node) => {
  const Children = ParentNode?.childNodes || [];
  const OrigChildDat = OrigNode?.childNodes || [];

  for (let i = 0; i < Children.length; i++) {
    const Child = Children[i];
    const nodeName = Child?.nodeName;

    if (CONTAINER_ELEMENTS.includes(nodeName)) {
      computeStyles(Child, OrigChildDat[i]);
    } else if (nodeName in RELEVANT_STYLES) {
      const styleDef = window.getComputedStyle(OrigChildDat[i] as Element);
      const styles = RELEVANT_STYLES[nodeName].reduce((stylesString, style) => {
        return `${stylesString} ${style}: ${styleDef.getPropertyValue(style)};`;
      }, '');
      (Child as Element).setAttribute('style', styles);
    }
  }
};

const downloadURLToFile = (url: string, projectName: string) => {
  const a = document.createElement('a');
  a.href = url;
  a.download = `${projectName}.svg`;
  document.body.appendChild(a);
  a.click();
  window.URL.revokeObjectURL(url);
  document.body.removeChild(a);
};

const downloadSVG = (
  width: number,
  height: number,
  projectName?: string,
  globalId?: string,
) =>
  new Promise((resolve, reject) => {
    try {
      const svgCopy = (document.getElementById('Svg')?.cloneNode(true) ||
        document.createElement('svg')) as Element;

      computeStyles(
        svgCopy,
        document.getElementById('Svg') || document.createElement('svg'),
      );

      // set useful attributes on the parent SVG
      svgCopy.setAttribute('height', String(height + LABEL_PADDING));
      svgCopy.setAttribute('width', String(width));
      svgCopy.setAttribute('viewBox', `0 0 ${width} ${height}`);
      svgCopy.setAttribute('style', 'background-color: #FAFAFA;');
      (svgCopy.childNodes[1] as Element).setAttribute(
        'transform',
        `translate(0, ${LABEL_PADDING})`,
      );

      // embed fonts
      const style = document.createElementNS(
        'http://www.w3.org/2000/svg',
        'style',
      );
      style.type = 'text/css';
      style.innerHTML = PublicFont;
      style.innerHTML += MontserratFont;
      (svgCopy.childNodes[0] as Element).appendChild(style);

      // add date and global id labels to top of SVG
      addLabel(
        svgCopy,
        `"${projectName}" ${format(new Date(), 'MM/dd/yyyy h:mmaaa OOO')}`,
      );
      globalId && addLabel(svgCopy, `Global ID ${globalId}`, 2);

      // export to file
      const data = new XMLSerializer().serializeToString(svgCopy);
      const svg = new Blob([data], {type: 'image/svg+xml;charset=utf-8'});
      const url = window.URL.createObjectURL(svg);

      downloadURLToFile(url, projectName || 'default_canvas');
      resolve(`${projectName} canvas downloaded`);
    } catch (e) {
      reject(e);
    }
  });

export default downloadSVG;
