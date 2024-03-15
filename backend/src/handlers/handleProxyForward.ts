import {Request, Response} from 'express';

export default async (req: Request, res: Response) => {
  const authToken = req.cookies.dashAuthToken;
  const isTest = process.env.NODE_ENV === 'test';
  const isDevelopment = process.env.NODE_ENV === 'development';

  const proxyPort = process.env.PACHYDERM_PROXY_SERVICE_PORT;

  try {
    const requestUrl = req.originalUrl.slice('/proxyForward'.length);
    let fullUrl = '';
    if (isDevelopment || isTest) {
      fullUrl = `http://localhost${requestUrl}?authn-token=${authToken}`;
    } else {
      // We are safe to assume requests coming from console in production
      // will be served under the same protocal and host as pachyderm

      // pachyderm-proxy is the name of the service that gets exposed in kubernetes
      if (req.hostname !== 'localhost' && req.hostname !== '127.0.0.1') {
        fullUrl = `${req.protocol}://${req.get(
          'host',
        )}${requestUrl}?authn-token=${authToken}`;
      } else {
        fullUrl = `${req.protocol}://pachyderm-proxy${
          proxyPort ? `:${proxyPort}` : ''
        }${requestUrl}?authn-token=${authToken}`;
      }
    }

    const {body, headers} = await fetch(fullUrl);

    if (!body) {
      throw new Error('fetch response body is null.');
    }

    body.pipeTo(
      new WritableStream({
        start() {
          headers.forEach((v, n) => res.setHeader(n, v));
        },
        write(chunk) {
          res.write(chunk);
        },
        close() {
          res.end();
        },
      }),
    );
  } catch (error) {
    return res.status(500).send(error);
  }
};
