import {cpp} from '@codemirror/lang-cpp';
import {css} from '@codemirror/lang-css';
import {html} from '@codemirror/lang-html';
import {java} from '@codemirror/lang-java';
import {javascript} from '@codemirror/lang-javascript';
import {json} from '@codemirror/lang-json';
import {markdown} from '@codemirror/lang-markdown';
import {php} from '@codemirror/lang-php';
import {python} from '@codemirror/lang-python';
import {rust} from '@codemirror/lang-rust';
import {sql} from '@codemirror/lang-sql';
import {xml} from '@codemirror/lang-xml';
import {LanguageSupport} from '@codemirror/language';

import parsePath from '@dash-frontend/lib/parsePath';

import {
  docker,
  go,
  julia,
  protobuf,
  r,
  ruby,
  shell,
  yaml,
} from '../extensions/legacy';

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

export type SupportedLanguagePlugins =
  | 'markdown'
  | 'yaml'
  | 'json'
  | 'python'
  | 'javascript'
  | 'html'
  | 'xml'
  | 'css'
  | 'cpp'
  | 'java'
  | 'php'
  | 'rust'
  | 'sql'
  | 'go'
  | 'shell'
  | 'r'
  | 'julia'
  | 'ruby'
  | 'docker'
  | 'protobuf';

export type SupportedFileIcons =
  | 'document'
  | 'image'
  | 'video'
  | 'audio'
  | 'folder'
  | 'unknown';

export type SupportedLanguage = SupportedLanguagePlugins | 'text';

type FileType = string;
type FileExtension = string;

type FileInfo = {
  extensions: FileExtension[];
  icon: SupportedFileIcons;
  renderer: SupportedRenderer;
  language?: SupportedLanguage;
  supportsPreview: boolean;
  supportsViewSource: boolean;
};

type FileTypeMap = Record<FileType, FileInfo>;
type FileExtensionMap = Record<FileExtension, FileInfo>;
type FilePluginMap = Record<SupportedLanguagePlugins, () => LanguageSupport>;

export const FILE_PLUGIN_MAP = Object.freeze<FilePluginMap>({
  markdown,
  yaml,
  json,
  python,
  javascript: () => javascript({jsx: true, typescript: true}),
  html,
  xml,
  css,
  cpp,
  java,
  php,
  rust,
  sql,
  go,
  shell,
  r,
  julia,
  ruby,
  docker,
  protobuf,
});

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
    language: 'html',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: true,
    extensions: ['.html', '.htm'],
  },

  xml: {
    renderer: 'iframe',
    language: 'xml',
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

  python: {
    renderer: 'code',
    language: 'python',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.py'],
  },

  javascript: {
    renderer: 'code',
    language: 'javascript',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.js', '.jsx', '.ts', '.tsx', '.cjs', '.mjs'],
  },

  cpp: {
    renderer: 'code',
    language: 'cpp',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.c', '.cpp'],
  },

  java: {
    renderer: 'code',
    language: 'java',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.java'],
  },

  php: {
    renderer: 'code',
    language: 'php',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.php'],
  },

  rust: {
    renderer: 'code',
    language: 'rust',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.rs'],
  },

  sql: {
    renderer: 'code',
    language: 'sql',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.sql'],
  },

  go: {
    renderer: 'code',
    language: 'go',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.go'],
  },

  shell: {
    renderer: 'code',
    language: 'shell',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.sh'],
  },

  r: {
    renderer: 'code',
    language: 'r',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.r'],
  },

  julia: {
    renderer: 'code',
    language: 'julia',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.jl'],
  },

  ruby: {
    renderer: 'code',
    language: 'ruby',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.rb'],
  },

  docker: {
    renderer: 'code',
    language: 'docker',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['dockerfile'],
  },

  css: {
    renderer: 'code',
    language: 'css',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.css'],
  },

  protobuf: {
    renderer: 'code',
    language: 'protobuf',
    icon: 'document',
    supportsPreview: true,
    supportsViewSource: false,
    extensions: ['.proto'],
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
    language: 'xml',
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
      extensionMap[extension.toLowerCase()] = fileType;
    });

    return extensionMap;
  }, {}),
);

export const parseFilePath = (filePath: string) => {
  return parsePath(filePath);
};

export const getFileDetails = (filePath: string) => {
  const parsedPath = parsePath(filePath);
  const extension = (parsedPath.ext || parsedPath.name).toLowerCase();

  return FILE_EXTENSION_MAP[extension] || FILE_TYPE_MAP.unknown;
};

export const getFileLanguagePlugin = (language?: SupportedLanguage) => {
  if (!language || language === 'text') {
    return;
  }

  return FILE_PLUGIN_MAP[language];
};
