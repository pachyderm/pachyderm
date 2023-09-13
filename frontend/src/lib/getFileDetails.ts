import parsePath from './parsePath';

export type SupportedRenderer =
  | 'code'
  | 'markdown'
  | 'web'
  | 'iframe'
  | 'csv'
  | 'image'
  | 'audio'
  | 'video'
  | 'unknown';

export type SupportedLanguage = 'markdown' | 'yaml' | 'json' | 'text';

type FileType = string;
type FileExtension = string;

export type FileIconType =
  | 'document'
  | 'image'
  | 'video'
  | 'audio'
  | 'folder'
  | 'unknown';

type FileInfo = {
  extensions: FileExtension[];
  icon: FileIconType;
  renderer: SupportedRenderer;
  language?: SupportedLanguage;
  supportsPreview: boolean;
  supportsViewSource: boolean;
};

type FileTypeMap = Record<FileType, FileInfo>;
type FileExtensionMap = Record<FileExtension, FileInfo>;

export const FILE_TYPE_MAP = Object.freeze<FileTypeMap>({
  markdown: {
    renderer: 'markdown',
    language: 'markdown',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: true,
    extensions: [
      '.markdn',
      '.markdown',
      '.md',
      '.mdown',
      '.mdwn',
      '.mkdn',
      '.mkdown',
      '.mkd',
    ],
  },

  html: {
    renderer: 'web',
    language: 'text', // TODO: add HTML CodeMirror plugin
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: true,
    extensions: ['.html', '.htm'],
  },

  xml: {
    renderer: 'iframe',
    language: 'text', // TODO: add XML CodeMirror plugin
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: true,
    extensions: ['.xml', '.xsl'],
  },

  yaml: {
    renderer: 'code',
    language: 'yaml',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.yaml', '.yml'],
  },

  json: {
    renderer: 'code',
    language: 'json',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.json', '.jsonl'],
  },

  text: {
    renderer: 'code',
    language: 'text',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.text', '.txt', '.textpb'],
  },

  csv: {
    renderer: 'csv',
    language: 'text',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: true,
    extensions: ['.csv', '.tsv', '.tab'],
  },

  image: {
    renderer: 'image',
    icon: 'image',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: [
      '.jpeg',
      '.jpg',
      '.jfif',
      '.jif',
      '.jpe',
      '.pjpg',
      '.png',
      '.apng',
      '.gif',
      '.avif',
      '.avifs',
      '.webp',
      '.bmp',
      '.ico',
      '.tiff',
      '.tif',
    ],
  },

  svg: {
    renderer: 'image',
    language: 'text', // TODO: add an XML CodeMirror plugin
    icon: 'image',
    supportsPreview: true,
    supportsViewSource: true,
    extensions: ['.svg', '.svgz'],
  },

  video: {
    renderer: 'video',
    icon: 'video',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: [
      '.mpg',
      '.mpeg',
      '.mpe',
      '.m1v',
      '.m2v',
      '.mpa',
      '.mp4',
      '.mp4v',
      '.mpg4',
      '.avi',
      '.wmv',
      '.mov',
      '.qt',
      '.rm',
      '.ra',
      '.ram',
      '.swf',
      '.flv',
      '.webm',
    ],
  },

  audio: {
    renderer: 'audio',
    icon: 'audio',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: [
      '.mp3',
      '.m2a',
      '.m3a',
      '.mp2',
      '.mp2a',
      '.mpga',
      '.wav',
      '.ogg',
      '.oga',
      '.spx',
    ],
  },

  pdf: {
    renderer: 'unknown', // TODO: Look into https://github.com/mozilla/pdf.js
    icon: 'document',
    supportsPreview: false,
    supportsViewSource: false,
    extensions: ['.pdf'],
  },

  xls: {
    renderer: 'unknown', // TODO: Look into xls renderers
    icon: 'document',
    supportsPreview: false,
    supportsViewSource: false,
    extensions: ['.xls', '.xlsx'],
  },

  doc: {
    renderer: 'unknown', // TODO: Look into doc renderers
    icon: 'document',
    supportsPreview: false,
    supportsViewSource: false,
    extensions: ['.doc', '.docx'],
  },

  unknown: {
    renderer: 'unknown',
    icon: 'unknown',
    supportsPreview: false,
    supportsViewSource: false,
    extensions: [],
  },
});

export const FILE_EXTENSION_MAP = Object.freeze<FileExtensionMap>(
  Object.keys(FILE_TYPE_MAP).reduce<FileExtensionMap>((extensionMap, type) => {
    const fileType = FILE_TYPE_MAP[type];

    fileType.extensions.forEach((extension) => {
      extensionMap[extension] = fileType;
    });

    return extensionMap;
  }, {}),
);

export const parseFilePath = (filePath: string) => {
  return parsePath(filePath);
};

export const getFileDetails = (filePath: string) => {
  const extension = parsePath(filePath).ext;

  return FILE_EXTENSION_MAP[extension] || FILE_TYPE_MAP.unknown;
};
