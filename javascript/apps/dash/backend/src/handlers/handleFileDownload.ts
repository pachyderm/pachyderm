import {pachydermClient} from '@pachyderm/node-pachyderm';
import {Request, Response} from 'express';
import {lookup} from 'mime-types';

import loggingPlugin from '@dash-backend/grpc/plugins/loggingPlugin';
import log from '@dash-backend/lib/log';
const {GRPC_SSL} = process.env;

const handleFileDownload = async (req: Request, res: Response) => {
  const authToken = req.cookies.dashAuthToken;
  const pachdAddress = process.env.PACHD_ADDRESS;
  const sameOrigin =
    req.get('origin') === undefined || process.env.NODE_ENV === 'development';

  if (!sameOrigin || !authToken || !pachdAddress) {
    return res
      .send('You do not have permission to access this file.')
      .status(401);
  }

  const {commitId, repoName, branchName} = req.params;
  const path = req.params['0'];
  let data;

  try {
    data = await pachydermClient({
      pachdAddress,
      authToken,
      plugins: [loggingPlugin(log)],
      ssl: GRPC_SSL === 'true',
    })
      .pfs()
      .getFile({
        commitId,
        path,
        branch: {name: branchName, repo: {name: repoName}},
      });
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
