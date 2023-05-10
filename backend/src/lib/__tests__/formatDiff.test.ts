import formatDiff from '../formatDiff';

const input = [
  {
    newFile: {
      file: {
        commit: {
          repo: {
            name: 'montage',
            type: 'user',
            project: {name: 'tutorial'},
          },
          id: '98137fbcb27443baa8e86ce0fc826241',
          branch: {
            repo: {
              name: 'montage',
              type: 'user',
              project: {name: 'tutorial'},
            },
            name: 'master',
          },
        },
        path: '/',
        datum: '',
      },
      fileType: 2,
      committed: {seconds: 1683654256, nanos: 158277000},
      sizeBytes: 1139103,
      hash: 'DDtaRMZ1tPG7VHdwo0DiMaKP2JMzjAE8O+ZA8VPCxU0=',
    },
    oldFile: {
      file: {
        commit: {
          repo: {
            name: 'montage',
            type: 'user',
            project: {name: 'tutorial'},
          },
          id: 'bf009df9859d4a6f8012866bfd9c23b6',
          branch: {
            repo: {
              name: 'montage',
              type: 'user',
              project: {name: 'tutorial'},
            },
            name: 'master',
          },
        },
        path: '/',
        datum: '',
      },
      fileType: 2,
      committed: {seconds: 1683654254, nanos: 948288000},
      sizeBytes: 572005,
      hash: '6q6r4pU5pEJ24Eyl4krJ1aOWvi0xEgZFrgOuEoVHj0M=',
    },
  },
  {
    newFile: {
      file: {
        commit: {
          repo: {
            name: 'montage',
            type: 'user',
            project: {name: 'tutorial'},
          },
          id: '98137fbcb27443baa8e86ce0fc826241',
          branch: {
            repo: {
              name: 'montage',
              type: 'user',
              project: {name: 'tutorial'},
            },
            name: 'master',
          },
        },
        path: '/montage.png',
        datum:
          '2f83ae6f5460c07c2f6210385d16edcd47ab2ece3ba1d4c8e763708d1efaffd5',
      },
      fileType: 1,
      committed: {seconds: 1683654256, nanos: 158277000},
      sizeBytes: 1139103,
      hash: 'TJ+W6UwklhOtiMJyowIXjjtsIjzEdd5JVmwudjTIM8M=',
    },
    oldFile: {
      file: {
        commit: {
          repo: {
            name: 'montage',
            type: 'user',
            project: {name: 'tutorial'},
          },
          id: 'bf009df9859d4a6f8012866bfd9c23b6',
          branch: {
            repo: {
              name: 'montage',
              type: 'user',
              project: {name: 'tutorial'},
            },
            name: 'master',
          },
        },
        path: '/montage.png',
        datum:
          '2f83ae6f5460c07c2f6210385d16edcd47ab2ece3ba1d4c8e763708d1efaffd5',
      },
      fileType: 1,
      committed: {seconds: 1683654254, nanos: 948288000},
      sizeBytes: 572005,
      hash: 'zsEK4A43v6/lBwgfOGUccaCEqLpW3lqJvfLKHT7eiSA=',
    },
  },
];

describe('formatDiff', () => {
  it('gives correct output', () => {
    const res = formatDiff(input);

    expect(res).toStrictEqual({
      diffTotals: {'/': 'UPDATED', '/montage.png': 'UPDATED'},
      diff: {
        size: 567098,
        sizeDisplay: '567.1 kB',
        filesAdded: {count: 0, sizeDelta: 0},
        filesUpdated: {count: 1, sizeDelta: 567098},
        filesDeleted: {count: 0, sizeDelta: 0},
      },
    });
  });
});
