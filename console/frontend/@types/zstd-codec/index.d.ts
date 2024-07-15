declare module 'zstd-codec' {
  export interface ZstdSimple {
    compress: (data: Uint8Array) => Uint8Array;
  }

  export interface Zstd {
    Simple: new () => ZstdSimple;
  }

  export interface ZstdCodecType {
    run: (cb: (zstd: Zstd) => void) => void;
  }

  export const ZstdCodec: ZstdCodecType;
}
