import {Request, Response} from 'express';
import {lookup} from 'mime-types';

import client from '@dash-backend/grpc/client';
import log from '@dash-backend/lib/log';

const handleFileDownload = async (req: Request, res: Response) => {
  const authToken = req.cookies.dashAuthToken;
  const pachdAddress = req.cookies.dashAddress;
  const sameOrigin = req.get('origin') === undefined;

  if (!sameOrigin || !authToken || !pachdAddress) {
    return res
      .send('You do not have permission to access this file.')
      .status(401);
  }

  const {commitId, repoName} = req.params;
  const path = req.params['0'];
  let data;

  try {
    data = await client({pachdAddress, authToken, log})
      .pfs()
      .getFile({commitId, path, repoName});
  } catch (err) {
    data = '';
  }

  if (data) {
    res.writeHead(200, {
      'Content-Type': lookup(path) || 'application/octet-stream',
      'Content-Length': data.length,
    });
    return res.end(data);
  } else {
    return res.send('File does not exist.').status(404);
  }
};

export default handleFileDownload;
