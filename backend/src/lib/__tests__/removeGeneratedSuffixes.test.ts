import removeGeneratedSuffixes from '../removeGeneratedSuffixes';

describe('removeGeneratedSuffixes', () => {
  it('should remove List and Map suffixes from pipeline specs', () => {
    const pipelineSpec = {
      transform: {
        image: 'dpokidov/imagemagick:7.1.0-23',
        cmdList: ['sh'],
        stdinList: [
          'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
        ],
      },
      input: {
        crossList: [
          {
            pfs: {
              name: 'edges',
              repo: 'edges',
              repoType: 'user',
              branch: 'master',
              glob: '/',
            },
          },
          {
            pfs: {
              name: 'images',
              repo: 'images',
              repoType: 'user',
              branch: 'master',
              glob: '/',
            },
          },
        ],
      },
      reprocessSpec: 'until_success',
    };

    const originalSpec = removeGeneratedSuffixes(pipelineSpec) as Record<
      string,
      Record<string, unknown>
    >;
    expect(originalSpec.transform.cmdList).toBeUndefined();
    expect(originalSpec.transform.cmd).toBeDefined();
    expect(originalSpec.transform.cmd).toHaveLength(1);
  });
});
