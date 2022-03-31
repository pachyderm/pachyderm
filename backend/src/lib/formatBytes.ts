import filesize from 'filesize';

const formatBytes = (bytes: number) =>
  filesize(bytes, {roundingMethod: 'ceil'});

export default formatBytes;
