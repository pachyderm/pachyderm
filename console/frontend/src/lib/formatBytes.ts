import filesize from 'filesize';

const formatBytes = (bytes: number | string | undefined) => {
  if (!bytes) return 'N/A';
  return filesize(typeof bytes === 'string' ? Number(bytes) : bytes, {
    roundingMethod: 'ceil',
  });
};

export default formatBytes;
