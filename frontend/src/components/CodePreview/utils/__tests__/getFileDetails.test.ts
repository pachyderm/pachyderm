import {
  parseFilePath,
  getFileDetails,
  FILE_EXTENSION_MAP,
  FILE_TYPE_MAP,
} from '../getFileDetails';

describe('lib/getFileDetails', () => {
  it('should return file info for unknown files', () => {
    const fileDetails = FILE_TYPE_MAP.unknown;

    expect(fileDetails.extensions).toHaveLength(0);
    expect(getFileDetails('path/to/file.zip')).toMatchObject(fileDetails);
  });

  it.each([
    'markdown',
    'html',
    'xml',
    'yaml',
    'json',
    'text',
    'csv',
    'image',
    'svg',
    'video',
    'audio',
    'pdf',
    'xls',
    'doc',
  ])('should return file details for %s files', (subtype) => {
    const fileDetails = FILE_TYPE_MAP[subtype];
    const numExtensions = fileDetails.extensions.length;

    expect(numExtensions).toBeGreaterThan(0);

    fileDetails.extensions.forEach((extension) => {
      expect(getFileDetails(`path/to/file${extension}`)).toMatchObject(
        fileDetails,
      );
    });

    expect.assertions(numExtensions + 1);
  });

  it('should map every file extension to file details', () => {
    const extensions = Object.keys(FILE_EXTENSION_MAP);
    const numExtensions = extensions.length;

    extensions.forEach((extension) => {
      const fileDetails = FILE_EXTENSION_MAP[extension];

      expect(fileDetails.extensions).toContain(extension);
    });

    expect.assertions(numExtensions);
  });

  it('should never repeat an extension', () => {
    const extensions: {[key: string]: number} = {};
    const numExtensions = Object.keys(FILE_EXTENSION_MAP).length;

    Object.keys(FILE_TYPE_MAP).forEach((subtype) => {
      const fileDetails = FILE_TYPE_MAP[subtype];

      fileDetails.extensions.forEach((extension) => {
        extensions[extension] = extensions[extension]
          ? extensions[extension] + 1
          : 1;
      });
    });

    Object.keys(extensions).forEach((extension) => {
      expect(extensions[extension]).toBe(1);
    });

    expect.assertions(numExtensions);
  });

  it('should parse a file path', () => {
    expect(parseFilePath('README.md')).toMatchObject({
      base: 'README.md',
      ext: '.md',
      name: 'README',
    });
    expect(parseFilePath('path/to/README.md')).toMatchObject({
      base: 'README.md',
      ext: '.md',
      name: 'README',
    });
    expect(parseFilePath('/path/to/README.md')).toMatchObject({
      base: 'README.md',
      ext: '.md',
      name: 'README',
    });
    expect(parseFilePath('path\\to\\README.md')).toMatchObject({
      base: 'README.md',
      ext: '.md',
      name: 'README',
    });
    expect(parseFilePath('path.with.dots/to/README.md')).toMatchObject({
      base: 'README.md',
      ext: '.md',
      name: 'README',
    });
    expect(parseFilePath('README')).toMatchObject({
      base: 'README',
      ext: '',
      name: 'README',
    });
    expect(parseFilePath('.env')).toMatchObject({
      base: '.env',
      ext: '',
      name: '.env',
    });
    expect(parseFilePath('data.tar.gz')).toMatchObject({
      base: 'data.tar.gz',
      ext: '.gz',
      name: 'data.tar',
    });
    expect(parseFilePath('')).toMatchObject({base: '', ext: '', name: ''});
    expect(parseFilePath('/')).toMatchObject({base: '', ext: '', name: ''});
    expect(parseFilePath('//')).toMatchObject({base: '', ext: '', name: ''});
    expect(parseFilePath('//path/to')).toMatchObject({
      base: '',
      ext: '',
      name: '',
    });
    expect(parseFilePath('///')).toMatchObject({base: '', ext: '', name: ''});
    expect(parseFilePath('C')).toMatchObject({base: 'C', ext: '', name: 'C'});
    expect(parseFilePath('C:\\')).toMatchObject({
      base: '',
      ext: '',
      name: '',
    });
    expect(parseFilePath('C:')).toMatchObject({
      base: '',
      ext: '',
      name: '',
    });
    expect(parseFilePath('../')).toMatchObject({
      base: '..',
      ext: '',
      name: '..',
    });
    expect(parseFilePath('./')).toMatchObject({base: '.', ext: '', name: '.'});
    expect(parseFilePath('.')).toMatchObject({base: '.', ext: '', name: '.'});
  });
});
