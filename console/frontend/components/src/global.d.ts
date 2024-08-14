declare module '*.module.css' {
  const styles: any; // eslint-disable-line @typescript-eslint/no-explicit-any
  export default styles;
}

declare module '*.svg' {
  import React = require('react');
  export const ReactComponent: React.FunctionComponent<
    React.SVGProps<SVGSVGElement>
  >;
  const src: string;
  export default src;
}
declare module '*.webp' {
  const src: string;
  export default src;
}

declare interface Window {
  analyticsInitialized?: boolean;
  clusterIdentified?: boolean;
  analyticsIdentified?: boolean;
}
