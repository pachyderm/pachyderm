import {appendBuffer} from '../handleEncodeArchiveUrl';

describe('appendBuffers', () => {
  it('should append buffers correctly', () => {
    const res = appendBuffer(
      new Uint8Array([1, 2, 3]),
      new Uint8Array([3, 2, 1]),
    );

    expect(res).toStrictEqual(new Uint8Array([1, 2, 3, 3, 2, 1]));
  });
});
