const {exec} = require('child_process');

const executeCommand = (command) => {
  return new Promise((res, rej) => {
    exec(command, (err, stdout) => {
      if (err) {
        rej(err);
      } else {
        res(stdout);
      }
    });
  });
};

const executeSynchronousCommands = async (commands) => {
  if (!commands || !commands.length) return Promise.resolve();
  await executeCommand(commands[0]);
  return executeSynchronousCommands(commands.slice(1));
};

const setup = async () => {
  const dags = 2;
  const pipelines = 10;
  const commits = 25;

  const pipelineCommands = [...new Array(dags).keys()].reduce(
    (commands, dagIndex) => {
      commands.push(`pachctl create repo repo-${dagIndex}`);
      commands.push(
        `printf '${JSON.stringify({
          pipeline: {
            name: `dag-${dagIndex}-pipeline-0`,
          },
          input: {
            pfs: {
              glob: '/*',
              repo: `repo-${dagIndex}`,
            },
          },
          transform: {
            cmd: ['python3', '/edges.py'],
            image: 'pachyderm/opencv',
          },
        })}' | pachctl create pipeline`,
      );
      commands.push(
        ...[...new Array(pipelines).keys()]
          .map((i) => i + 1)
          .map(
            (pipelineIndex) =>
              `printf '${JSON.stringify({
                pipeline: {
                  name: `dag-${dagIndex}-pipeline-${pipelineIndex}`,
                },
                input: {
                  cross: [
                    {
                      pfs: {
                        glob: '/*',
                        repo: `repo-${dagIndex}`,
                      },
                    },
                    {
                      pfs: {
                        glob: '/*',
                        repo: `dag-${dagIndex}-pipeline-${pipelineIndex - 1}`,
                      },
                    },
                  ],
                },
                transform: {
                  cmd: ['sh'],
                  image: 'dpokidov/imagemagick:7.0.10-58',
                  stdin: [
                    'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
                  ],
                },
              })}' | pachctl create pipeline`,
          ),
      );
      return commands;
    },
    [],
  );

  console.log('building pipelines...');
  await executeSynchronousCommands(pipelineCommands);

  const commitCommands = [...new Array(commits).keys()].map(
    (commitIndex) =>
      `pachctl put file repo-${
        commitIndex % 2
      }@master:liberty-${commitIndex}.png -f ./etc/testing/files/46Q8nDz.jpg`,
  );
  console.log('committing files...');

  await executeSynchronousCommands(commitCommands);
};

setup();
