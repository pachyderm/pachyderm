/* eslint-disable @typescript-eslint/no-explicit-any */
declare module '*.module.css' {
  const styles: any;
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

declare interface Window {
  REACT_APP_OAUTH_URL: string | undefined;
  REACT_APP_OAUTH_CLIENT_ID: string | undefined;
  REACT_APP_OAUTH_PACHD_CLIENT_ID: string | undefined;
}
