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
    }
  }
}

export {};
