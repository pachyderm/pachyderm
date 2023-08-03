import {ApolloError} from 'apollo-server-errors';
import {Request, Response} from 'express';
import {lookup} from 'mime-types';

import getPachClientAndAttachHeaders from '@dash-backend/lib/getPachClient';

const handleFileDownload = async (req: Request, res: Response) => {
  const authToken = req.cookies.dashAuthToken;
  const pachdAddress = process.env.PACHD_ADDRESS;
  const isTest = process.env.NODE_ENV === 'test';
  const sameOrigin =
    req.get('origin') === undefined ||
    process.env.NODE_ENV === 'development' ||
    isTest;

  if (!isTest && (!sameOrigin || !pachdAddress)) {
    return res
      .send('You do not have permission to access this file.')
      .status(401);
  }

  const pachClient = getPachClientAndAttachHeaders({
    requestId: req.header('x-request-id') || '',
    authToken,
  });

  if (!isTest && !authToken && pachdAddress) {
    try {
      await pachClient.auth.whoAmI();
    } catch (e) {
      const {code} = (e as ApolloError).extensions;
      if (code !== 'UNIMPLEMENTED') {
        return res
          .send('You do not have permission to access this file.')
          .status(401);
      }
    }
  }

  const {commitId, repoName, branchName, projectId} = req.params;
  const path = req.params['0'];
  let data;

  try {
    const pachClient = getPachClientAndAttachHeaders({
      requestId: req.header('x-request-id') || '',
      authToken,
      projectId,
    });
    data = await pachClient.pfs.getFile({
      projectId,
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
