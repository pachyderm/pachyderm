import {ZstdCodec} from 'zstd-codec';

export const zstdCompress = (encodedData: Uint8Array) => {
  return new Promise<Uint8Array>((resolve) => {
    ZstdCodec.run((zstd) => {
      const simple = new zstd.Simple();
      const compressedUrl: Uint8Array = simple.compress(encodedData);
      resolve(compressedUrl);
    });
  });
};
