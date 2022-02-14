import {ServiceError} from '@grpc/grpc-js';
import {
  Commit,
  pachydermClient,
  STREAM_OVERHEAD_LENGTH,
} from '@pachyderm/node-pachyderm';
import {ModifyFile} from '@pachyderm/node-pachyderm/dist/services/pfs/clients/ModifyFile';
import busboy from 'busboy';
import {Request, Response} from 'express';

import errorPlugin from '@dash-backend/grpc/plugins/errorPlugin';
import loggingPlugin from '@dash-backend/grpc/plugins/loggingPlugin';
import {GRPC_MAX_MESSAGE_LENGTH} from '@dash-backend/lib/constants';
import log from '@dash-backend/lib/log';
const {GRPC_SSL, PACHD_ADDRESS} = process.env;

type FileUploadBody = {
  path: string;
  commit: Commit.AsObject;
  chunk: Blob;
};

const BUSBOY_CHUNK_SIZE = 1024 * 64; // 64 KiB

const handleFileUpload = async (
  req: Request<Request['params'], unknown, FileUploadBody>,
  res: Response,
) => {
  let fileClient: ModifyFile | undefined;

  const uploadFields = {
    path: '',
    branch: '',
    repo: '',
  };

  const authToken = req.cookies.dashAuthToken;

  if (!authToken || !PACHD_ADDRESS) {
    return res
      .send('You do not have permission to access this file.')
      .status(401);
  }

  const pachClient = pachydermClient({
    pachdAddress: PACHD_ADDRESS,
    authToken,
    plugins: [loggingPlugin(log), errorPlugin],
    ssl: GRPC_SSL === 'true',
  });

  const bb = busboy({headers: req.headers});

  bb.on('error', (err: string) => {
    req.unpipe(bb);

    fileClient && fileClient.stream.destroy(new Error(err));
    res.send(err).status(500);
  });

  bb.on('field', (name, value, _info) => {
    uploadFields[name as keyof typeof uploadFields] = value;
  }).on('file', async (_name, file, _info) => {
    fileClient = await pachClient.pfs().modifyFile();
    fileClient.autoCommit({
      name: uploadFields.branch,
      repo: {name: uploadFields.repo},
    });

    const chunkLimit = Math.floor(
      (GRPC_MAX_MESSAGE_LENGTH -
        STREAM_OVERHEAD_LENGTH -
        uploadFields.path.length) /
        BUSBOY_CHUNK_SIZE,
    );

    // eslint-disable-next-line  @typescript-eslint/no-explicit-any
    let chunks: any[] = [];

    file.on('end', () => {
      if (chunks.length && fileClient) {
        fileClient.putFileFromBytes(uploadFields.path, Buffer.concat(chunks));
        chunks = [];
      }
    });

    file.on('data', (chunk) => {
      chunks.push(chunk); // buffer chunks until we hit the limit for a GRPC stream so that we don't send a bunch of small writes that overflow the GRPC streams buffer
      if (chunks.length === chunkLimit && fileClient) {
        fileClient.putFileFromBytes(uploadFields.path, Buffer.concat(chunks));
        chunks = [];
      }
    });
  });

  req.on('close', () => {
    fileClient &&
      fileClient.stream.destroy(new Error('Upload cancelled by client'));

    req.unpipe(bb);
  });

  bb.on('close', async () => {
    try {
      if (fileClient) await fileClient.end();
      res.send({data: 'upload successful'});
    } catch (err) {
      res.send((err as ServiceError).message).status(500);
    }
  });

  req.pipe(bb);
};

export {handleFileUpload};
