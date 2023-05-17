// Creates a new Uint8Array based on two different ArrayBuffers
export const appendBuffer = (buffer1: Uint8Array, buffer2: Uint8Array) => {
  const tmp = new Uint8Array(buffer1.byteLength + buffer2.byteLength);
  tmp.set(buffer1, 0);
  tmp.set(buffer2, buffer1.byteLength);
  return tmp;
};
