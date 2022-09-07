import {FileMajorType} from './types';

const getFileMajorType = (fileType: string) => {
  let fileMajorType: FileMajorType;

  switch (fileType) {
    case 'pdf':
    case 'xls':
    case 'xlsx':
    case 'html':
    case 'doc':
    case 'docx':
    case 'md':
    case 'csv':
    case 'json':
    case 'jsonl':
    case 'yml':
    case 'yaml':
    case 'txt':
    case 'textpb':
      fileMajorType = 'document';
      break;
    case 'apng':
    case 'avif':
    case 'gif':
    case 'jpeg':
    case 'jpg':
    case 'png':
    case 'svg':
    case 'webp':
    case 'bmp':
    case 'ico':
    case 'tiff':
      fileMajorType = 'image';
      break;
    case 'mpg':
    case 'mpeg':
    case 'avi':
    case 'wmv':
    case 'mov':
    case 'rm':
    case 'ram':
    case 'swf':
    case 'flv':
    case 'pff':
    case 'webm':
    case 'mp4':
      fileMajorType = 'video';
      break;
    case 'mp3':
    case 'wav':
    case 'ogg':
      fileMajorType = 'audio';
      break;
    case 'folder':
      fileMajorType = 'folder';
      break;
    default:
      fileMajorType = 'unknown';
      break;
  }

  return fileMajorType;
};

export default getFileMajorType;
