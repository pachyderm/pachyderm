import filesize from 'filesize';

const formatBytes = (bytes: number) => filesize(bytes);

export default formatBytes;
