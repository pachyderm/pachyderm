declare module '*.module.css' {
  const styles: any;
  export default styles;
}

declare module '*.svg' {
  import React = require('react');
  export const ReactComponent: React.FunctionComponent<React.SVGProps<
    SVGSVGElement
  >>;
  const src: string;
  export default src;
}
