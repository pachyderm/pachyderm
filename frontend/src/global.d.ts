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

declare global {
  namespace NodeJS {
    interface ProcessEnv {
      [key: string]: string;
      ISSUER_URI: string;
      OAUTH_REDIRECT_URI: string;
      OAUTH_CLIENT_ID: string;
      OAUTH_CLIENT_SECRET: string;
      OAUTH_PACHD_CLIENT_ID: string;
      PACHD_ADDRESS: string;
      GRPC_SSL: string;
    }
  }
}
