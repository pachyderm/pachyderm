import {Request, Response} from 'express';
import {ZstdCodec} from 'zstd-codec';

interface ArchiveRequest extends Request {
  body: {
    projectId: string;
    repoId: string;
    commitId: string;
    paths: string[];
  };
}

export const FILE_DOWNLOAD_API_VERSION = 1;

export const encodeArchiveUrl = async (req: ArchiveRequest, res: Response) => {
  try {
    const {projectId, repoId, commitId, paths} = req.body;

    // 1. Sort the list of paths asciibetically.
    const sortedPaths = paths
      .map((p) => `${projectId}/${repoId}@${commitId}:${p}`)
      .sort();
    // 2. Terminate each file name with a `0x00` byte and concatenate to one string. A request MUST end with a `0x00` byte.
    const terminatedString = sortedPaths.join('\x00');
    // 2. a. encode to Uint8Array
    const encoder = new TextEncoder();
    const encodedData = encoder.encode(terminatedString);

    // 3. Compress the result with [Zstandard](https://facebook.github.io/zstd/).
    const compressedData = await zstdCompress(encodedData);

    // 4. Prefix the result with `0x01` (for URL format version 1).
    const prefixedData = appendBuffer(
      new Uint8Array([FILE_DOWNLOAD_API_VERSION]),
      compressedData,
    );

    // 5. Convert the result to URL-safe Base64.  See section 3.5 of RFC4648 for a description of this coding.  Padding may be omitted.
    const base64url = Buffer.from(prefixedData).toString('base64url');

    // 6. Prefix the result with `/archive` and suffix it with the desired archive format, like `.tar.zst`
    return res.json({url: `/proxyForward/archive/${base64url}.zip`});
  } catch (error) {
    return res
      .json({
        message: 'Error encoding archive url',
        details: [String(error)],
      })
      .status(500);
  }
};

// Creates a new Uint8Array based on two different ArrayBuffers
export const appendBuffer = (buffer1: Uint8Array, buffer2: Uint8Array) => {
  const tmp = new Uint8Array(buffer1.byteLength + buffer2.byteLength);
  tmp.set(buffer1, 0);
  tmp.set(buffer2, buffer1.byteLength);
  return tmp;
};

export const zstdCompress = (encodedData: Uint8Array) => {
  return new Promise<Uint8Array>((resolve) => {
    ZstdCodec.run((zstd) => {
      const simple = new zstd.Simple();
      const compressedUrl: Uint8Array = simple.compress(encodedData);
      resolve(compressedUrl);
    });
  });
};
