import {ClientReadableStream} from '@grpc/grpc-js';

const streamToObjectArray = <T extends {toObject: () => U}, U>(
  stream: ClientReadableStream<T>,
  limit = 0,
) => {
  return new Promise<U[]>((resolve, reject) => {
    const data: U[] = [];

    stream.on('data', (chunk: T) => {
      data.push(chunk.toObject());

      if (limit && data.length >= limit) {
        stream.destroy();
      }
    });
    stream.on('end', () => resolve(data));
    stream.on('close', () => resolve(data));
    stream.on('error', (err) => reject(err));
  });
};

export default streamToObjectArray;
