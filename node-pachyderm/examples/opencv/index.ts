import {join} from 'path';

import {pachydermClient} from '../../dist';
const pachdAddress = 'localhost:30650';
const ssl = false;

const main = async () => {
  try {
    // Connect to a pachyderm cluster on the default host:port
    const client = pachydermClient({pachdAddress, ssl});

    // Create images repo
    console.log('Creating images repo.');
    await client.pfs().createRepo({repo: {name: 'images'}});

    // Create edges pipeline
    console.log('Creating edges pipeline.');
    await client.pps().createPipeline({
      pipeline: {
        name: 'edges',
      },
      transform: {
        image: 'pachyderm/opencv',
        cmdList: ['python3', '/edges.py'],
      },
      input: {
        pfs: {
          repo: 'images',
          glob: '/*',
          branch: 'master',
        },
      },
    });

    // Create montage pipeline
    console.log('Creating montage pipeline.');
    await client.pps().createPipeline({
      pipeline: {
        name: 'montage',
      },
      transform: {
        image: 'v4tech/imagemagick',
        cmdList: ['sh'],
        stdinList: [
          'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
        ],
      },
      input: {
        crossList: [
          {
            pfs: {
              repo: 'images',
              glob: '/',
              branch: 'master',
            },
          },
          {
            pfs: {
              repo: 'edges',
              glob: '/',
              branch: 'master',
            },
          },
        ],
      },
    });

    // Add some images from the images directory. Alternatively, you could use client.putFileFromURL or
    //  client.putFileFromBytes.
    console.log('Adding files from images directory to the images repo.');
    const commit = await client
      .pfs()
      .startCommit({branch: {name: 'master', repo: {name: 'images'}}});
    const file = await client.pfs().modifyFile();
    await file
      .setCommit(commit)
      .putFileFromFilepath(join(__dirname, 'images/8MN9Kg0.jpg'), '8MN9Kg0.jpg')
      .putFileFromFilepath(join(__dirname, 'images/46Q8nDz.jpg'), '46Q8nDz.jpg')
      .putFileFromFilepath(join(__dirname, 'images/g2QnNqa.jpg'), 'g2QnNqa.jpg')
      .end();

    await client.pfs().finishCommit({commit});

    // Wait for the commit (and its downstream commits) to finish
    console.log('Waiting for commits to finish.');
    await client.pfs().inspectCommitSet({wait: true, commitSet: commit});

    console.log('Montage created.');
  } catch (e) {
    console.log(e);
  }
};

main();
