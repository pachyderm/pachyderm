import {FileInfo, FileType} from '@dash-backend/proto';
import {fileInfoFromObject} from '@dash-backend/proto/builders/pfs';

import {FILES} from './loadLimits';

const tutorial = {
  '/': [
    fileInfoFromObject({
      committed: {seconds: 1614126189, nanos: 0},
      file: {
        commitId: 'd350c8d08a644ed5b2ee98c035ab6b33',
        path: '/AT-AT.png',
        branch: {name: 'master', repo: {name: 'images'}},
      },
      fileType: FileType.FILE,
      hash: 'P2fxZjakvux5dNsEfc0iCx1n3Kzo2QcDlzu9y3Ra1gc=',
      sizeBytes: 80588,
    }),
    fileInfoFromObject({
      committed: {seconds: 1610126189, nanos: 0},
      file: {
        commitId: 'd350c8d08a644ed5b2ee98c035ab6b33',
        path: '/liberty.png',
        branch: {name: 'master', repo: {name: 'images'}},
      },
      fileType: FileType.FILE,
      hash: 'QJoij3vijagMPZAadaQ1PtLRL7NFNZgouPoPQLbSi8E=',
      sizeBytes: 58644,
    }),
    fileInfoFromObject({
      committed: {seconds: 1616126189, nanos: 0},
      file: {
        commitId: 'd350c8d08a644ed5b2ee98c035ab6b33',
        path: '/cats/',
        branch: {name: 'master', repo: {name: 'images'}},
      },
      fileType: FileType.DIR,
      hash: 'sMAJnz5xBEFfYUVUo5PZrOpaoPZ902b+7N6+Fg5ACkQ=',
      sizeBytes: 98747,
    }),
  ],
  '/cats/': [
    fileInfoFromObject({
      committed: {seconds: 1612126189, nanos: 0},
      file: {
        commitId: 'd350c8d08a644ed5b2ee98c035ab6b33',
        path: '/cats/kitten.png',
        branch: {name: 'master', repo: {name: 'images'}},
      },
      fileType: FileType.FILE,
      hash: 'AGyiZfGAxLyuqfB1yCGe/AMCpp2zdTYW8B37j8ls4hA',
      sizeBytes: 104836,
    }),
  ],
};

const allFiles = {
  ...tutorial,
  '/': [
    ...tutorial['/'],
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/carriers_list.textpb',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 41700,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/csv_commas.csv',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 146,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/csv_tabs.csv',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 146,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/html_pachyderm.html',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 110959,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/json_mixed.json',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 2896,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/json_nested_arrays.json',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 2839,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/json_object_array.json',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 2232,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/json_single_field.json',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 58,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/json_string_array.json',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 33750,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/jsonl_people.jsonl',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 214,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/tsv_tabs.tsv',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 146,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/txt_spec.txt',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 502,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/xml_plants.xml',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 8020,
    }),
    fileInfoFromObject({
      committed: {seconds: 1633119338, nanos: 0},
      file: {
        commitId: '531f844bd184e913b050d49856e8d438',
        path: '/yml_spec.yml',
        branch: {name: 'master', repo: {name: 'samples'}},
      },
      fileType: FileType.FILE,
      hash: '1aa5784d52481911bc44df0e8b6a8fd581b0518c',
      sizeBytes: 502,
    }),
  ],
};

const nestedFolders = (() => {
  const sampleWords =
    `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum`.split(
      ' ',
    );
  let path = '/';
  return sampleWords.reduce((memo: {[key: string]: FileInfo[]}, word) => {
    const nextPath = path + word + '/';
    memo[path] = [
      fileInfoFromObject({
        committed: {seconds: 1616126189, nanos: 0},
        file: {
          commitId: 'd350c8d08a644ed5b2ee98c035ab6b34',
          path: nextPath,
          branch: {name: 'master', repo: {name: 'images'}},
        },
        fileType: FileType.DIR,
        hash: 'sMAJnz5xBEFfYUVUo5PZrOpaoPZ902b+7N6+Fg5ACkQ=',
        sizeBytes: 98747,
      }),
    ];

    path = nextPath;
    return memo;
  }, {});
})();

const getLoadFiles = (fileCount: number) => {
  const now = Math.floor(new Date().getTime() / 1000);
  return {
    '/': [...new Array(fileCount).keys()].map((fileIndex) => {
      return fileInfoFromObject({
        committed: {seconds: now - 100 * fileIndex, nanos: 0},
        file: {
          commitId: `${0}-${0}`,
          path: `${fileIndex}.json`,
          branch: {name: 'master', repo: {name: 'images'}},
        },
        fileType: FileType.FILE,
        hash: '',
        sizeBytes: Math.floor(Math.random() * 1000),
      });
    }),
  };
};

export type Files = {
  [projectId: string]: {
    [path: string]: FileInfo[];
  };
};

const files: Files = {
  '1': allFiles,
  '2': tutorial,
  '3': allFiles,
  '4': tutorial,
  '5': nestedFolders,
  '9': getLoadFiles(FILES),
  default: tutorial,
};

export default files;
