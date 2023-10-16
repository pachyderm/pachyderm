import {ServiceError} from '@grpc/grpc-js';
import {ApolloError} from 'apollo-server-errors';
import busboy from 'busboy';
import cors from 'cors';
import {NextFunction, Request, Response, Router} from 'express';
import PQueue from 'p-queue';

import isServiceError from '@dash-backend/grpc/utils/isServiceError';
import {GRPC_MAX_MESSAGE_LENGTH} from '@dash-backend/lib/constants';
import FileUploads, {upload} from '@dash-backend/lib/FileUploads';
import getPachClientAndAttachHeaders from '@dash-backend/lib/getPachClient';
import baseLogger from '@dash-backend/lib/log';
import {Commit, STREAM_OVERHEAD_LENGTH} from '@dash-backend/proto';
import {FileSet} from '@dash-backend/proto/services/pfs/clients/FileSet';

const {PACHD_ADDRESS} = process.env;

type FileUploadBody = {
  deleteFile: boolean;
  currentChunk: number;
  totalChunks: number;
  uploadId: string;
  fileName: string;
};

type FileUploadStartBody = {
  path: string;
  repo: string;
  branch: string;
  description: string;
  projectId: string;
};

const BUSBOY_CHUNK_SIZE = 1024 * 64; // 64 KiB

const isTest = process.env.NODE_ENV === 'test';

const logError = (error: unknown, requestName: string) => {
  if (isServiceError(error)) {
    baseLogger.error({error: error.details}, `${requestName} request failed`);
  } else {
    baseLogger.error({error}, `${requestName} request failed`);
  }
};

const uploadStart = async (
  req: Request<Request['params'], unknown, FileUploadStartBody>,
  res: Response,
  next: NextFunction,
) => {
  const authToken = req.cookies.dashAuthToken;
  const {path, branch, repo, description, projectId} = req.body;

  if (!PACHD_ADDRESS) {
    res.status(401);
    next('You do not have permission to upload files');
  }

  if (!isTest && !authToken && PACHD_ADDRESS) {
    try {
      const pachClient = getPachClientAndAttachHeaders({
        requestId: req.header('x-request-id') || '',
        authToken,
        projectId,
      });
      await pachClient.auth.whoAmI();
    } catch (e) {
      const {code} = (e as ApolloError).extensions;
      if (code !== 'UNIMPLEMENTED') {
        res.status(401);
        next('You do not have permission to upload files');
      }
    }
  }

  if (!repo) {
    res.status(400);
    next('Repo name must be specified');
  }

  if (!branch) {
    res.status(400);
    next('Branch name must be specified');
  }

  if (!path) {
    res.status(400);
    next('Path must be specified');
  }

  const id = FileUploads.addUpload(path, repo, branch, description, projectId);

  return res.send({uploadId: id});
};

const uploadFinish = async (
  req: Request<Request['params'], unknown, {uploadId: string}>,
  res: Response,
  next: NextFunction,
) => {
  const authToken = req.cookies.dashAuthToken;

  if (!PACHD_ADDRESS) {
    res.status(401);
    next('You do not have permission to upload files');
  }

  const pachClient = getPachClientAndAttachHeaders({
    requestId: req.header('x-request-id') || '',
    authToken,
  });

  if (!isTest && !authToken) {
    try {
      await pachClient.auth.whoAmI();
    } catch (e) {
      const {code} = (e as ApolloError).extensions;
      if (code !== 'UNIMPLEMENTED') {
        res.status(401);
        next('You do not have permission to upload files');
      }
    }
  }

  let commit: Commit.AsObject | undefined;
  const upload = FileUploads.getUpload(req.body.uploadId);
  if (upload) {
    try {
      const composedFileSet = await pachClient.pfs.composeFileSet(
        FileUploads.getFileSetIds(req.body.uploadId),
      );
      commit = await pachClient.pfs.startCommit({
        projectId: upload.projectId,
        branch: {
          name: upload.branch,
          repo: {
            name: upload.repo,
          },
        },
        description: upload.description,
      });
      await pachClient.pfs.addFileSet({
        projectId: upload.projectId,
        fileSetId: composedFileSet,
        commit,
      });
      await pachClient.pfs.finishCommit({projectId: upload.projectId, commit});
      FileUploads.removeUpload(req.body.uploadId);
      return res.send({commitId: commit.id});
    } catch (err) {
      logError(err, '');
      if (commit) {
        try {
          await pachClient.pfs.dropCommitSet({id: commit.id});
        } catch (err) {
          logError(err, 'dropCommitSet');
        }
      }

      res.status(500);
      next(err);
    }
  } else {
    res.status(404);
    next('Upload not found');
  }
};

const uploadFileCancel = async (
  req: Request<Request['params'], unknown, {fileId: string; uploadId: string}>,
  res: Response,
  next: NextFunction,
) => {
  const {fileId, uploadId} = req.body;

  if (!fileId) {
    res.status(404);
    next('File not found');
  }

  if (!uploadId) {
    res.status(404);
    next('Upload not found');
  }

  if (FileUploads.getUpload(uploadId)) {
    FileUploads.removeFile(fileId, uploadId);
    return res.send({fileId});
  } else {
    res.status(404);
    next('Upload not found');
  }
};

const uploadChunk = async (
  req: Request<Request['params'], unknown, FileUploadBody>,
  res: Response,
  next: NextFunction,
) => {
  let fileClient: FileSet;
  let uploadId: string;
  let path: string;
  let currentChunk: number;
  let chunkTotal: number;
  let upload: upload;
  let fileName = '';

  const authToken = req.cookies.dashAuthToken;

  if (!PACHD_ADDRESS) {
    return res.status(401).send('You do not have permission to upload files');
  }

  const pachClient = getPachClientAndAttachHeaders({
    requestId: req.header('x-request-id') || '',
    authToken,
  });

  if (!isTest && !authToken) {
    try {
      await pachClient.auth.whoAmI();
    } catch (e) {
      const {code} = (e as ApolloError).extensions;
      if (code !== 'UNIMPLEMENTED') {
        return res
          .status(401)
          .send('You do not have permission to upload files');
      }
    }
  }

  const bb = busboy({headers: req.headers});
  const workQueue = new PQueue({concurrency: 1});

  async function handleError(fn: () => void) {
    workQueue.add(async () => {
      try {
        await fn();
      } catch (e) {
        bb.removeAllListeners();
        req.unpipe(bb);
        workQueue.pause();
        next(e);
        return e;
      }
    });
  }

  req.on('close', async () => {
    req.unpipe(bb);
    workQueue.pause();
    bb.removeAllListeners();
    if (fileClient) {
      try {
        await fileClient.end();
        FileUploads.removeFile(fileName, uploadId);
      } catch (err) {
        next('Connection Error');
      }
    }
  });

  bb.on('error', async (err: string) => {
    logError(err, '');
    bb.removeAllListeners();
    req.unpipe(bb);

    res.status(500).send(err);
  });

  bb.on('field', async (name, value, _info) => {
    await handleError(() => {
      if (name === 'currentChunk') currentChunk = Number(value);
      if (name === 'chunkTotal') chunkTotal = Number(value);
      if (name === 'uploadId') uploadId = value;
      if (name === 'fileName') fileName = value;

      upload = FileUploads.getUpload(uploadId);
      if (upload) {
        path = upload.path;
      } else {
        res.status(404);
        throw new Error('Upload not found');
      }
    });
  }).on('file', async (_name, file, _info) => {
    let dataWritten = false;
    await handleError(async () => {
      if (!currentChunk || !chunkTotal) {
        res.status(400);
        throw new Error('Chunk values must be specified');
      }

      const fullPath = `${path}/${fileName}`;
      const chunkLimit = Math.floor(
        (GRPC_MAX_MESSAGE_LENGTH - STREAM_OVERHEAD_LENGTH - fullPath.length) /
          BUSBOY_CHUNK_SIZE,
      );

      // eslint-disable-next-line  @typescript-eslint/no-explicit-any
      let chunks: any[] = [];

      try {
        fileClient = await pachClient.pfs.fileSet();
      } catch (err) {
        logError(err, 'fileSet');
        const message = (err as ServiceError).message;
        res.status(500);
        throw new Error(message);
      }

      file.on('end', async () => {
        if (fileClient) {
          if (chunks.length) {
            fileClient.putFileFromBytes(
              fullPath,
              Buffer.concat(chunks),
              currentChunk !== 1 || dataWritten,
            );
            dataWritten = true;
            chunks = [];
          }
        }
      });

      file.on('data', (chunk) => {
        chunks.push(chunk); // buffer chunks until we hit the limit for a GRPC stream so that we don't send a bunch of small writes that overflow the GRPC streams buffer
        if (chunks.length === chunkLimit && fileClient) {
          fileClient.putFileFromBytes(
            fullPath,
            Buffer.concat(chunks),
            currentChunk !== 1 || dataWritten,
          );
          dataWritten = true;
          chunks = [];
        }
      });
    });
  });

  bb.on('finish', async () => {
    return handleError(async () => {
      try {
        if (fileClient) {
          const fileSetId = await fileClient.end();
          if (currentChunk === 1) {
            FileUploads.addFile(fileName, uploadId, fileSetId);
          } else if (
            upload.fileSets[fileName] &&
            !upload.fileSets[fileName].deleted
          ) {
            FileUploads.addFile(
              fileName,
              uploadId,
              await pachClient.pfs.composeFileSet([
                upload.fileSets[fileName].id,
                fileSetId,
              ]),
            );
          }
          if (upload.expiration - Date.now() <= 60000) {
            await Promise.all(
              FileUploads.getFileSetIds(uploadId).map((fileSetId) =>
                pachClient.pfs.renewFileSet({
                  fileSetId,
                  duration: 1800, // 30 minutes
                }),
              ),
            );
            FileUploads.extendUploadExpiration(uploadId);
          } else if (
            chunkTotal === currentChunk &&
            upload.fileSets[fileName] &&
            !upload.fileSets[fileName].deleted
          ) {
            await pachClient.pfs.renewFileSet({
              fileSetId: upload.fileSets[fileName].id,
              duration: 1800,
            });
          }
        }
        return res.send({fileName});
      } catch (err) {
        logError(err, 'fileSet');

        const message = (err as ServiceError).message;

        res.status(500);
        throw new Error(message);
      }
    });
  });

  req.pipe(bb);
};

const uploadsRouter = Router();

const noCors =
  process.env.NODE_ENV === 'development' || process.env.NODE_ENV === 'test';

uploadsRouter.post('/', noCors ? [] : cors(), uploadChunk);
uploadsRouter.post('/start', noCors ? [] : cors(), uploadStart);
uploadsRouter.post('/finish', noCors ? [] : cors(), uploadFinish);
uploadsRouter.post('/cancel', noCors ? [] : cors(), uploadFileCancel);

export default uploadsRouter;
