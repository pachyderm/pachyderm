import {parseArgs, promisify} from 'node:util';
import {exec} from 'node:child_process';
import range from 'lodash/range';

const execPromise = promisify(exec);
const {values} = parseArgs({
  options: {
    distributed: {type: 'boolean'},
    mode: {type: 'string'},
    numProjects: {type: 'string', default: '10'},
    numInaccessibleProjects: {type: 'string', default: '0'},
    numRepos: {type: 'string', default: '10'},
    numPipelines: {type: 'string', default: '10'},
    numCommits: {type: 'string', default: '10'},
    numFiles: {type: 'string', default: '100'},
    sizeFileSmall: {type: 'string', default: '1000'}, //~1KB
    sizeFileLarge: {type: 'string', default: '10000000'}, //~10MB
    sizeLogsLarge: {type: 'string', default: '1000000'}, //~1MB
    sizePipelineWorkersSmall: {type: 'string', default: '1'},
    sizePipelineCPUSmall: {type: 'string', default: '0.01'},
    sizePipelineCPULarge: {type: 'string', default: '1'},
    sizePipelineMemorySmall: {type: 'string', default: '100M'},
    sizePipelineMemoryLarge: {type: 'string', default: '1G'},
  },
});

const MODE = String(values.mode);
const DISTRIBUTED = Boolean(values.distributed);
const NUM_PROJECTS = Number(values.numProjects);
const NUM_INACCESSIBLE_PROJECTS = Number(values.numInaccessibleProjects);
const NUM_REPOS = Number(values.numRepos);
const NUM_PIPELINES = Number(values.numPipelines);
const NUM_COMMITS = Number(values.numCommits);
const NUM_FILES = Number(values.numFiles);
const SIZE_FILE_LARGE = Number(values.sizeFileLarge);
const SIZE_FILE_SMALL = Number(values.sizeFileSmall);
const SIZE_LOGS_LARGE = Number(values.sizeLogsLarge);
const SIZE_PIPELINE_WORKERS_SMALL = Number(values.sizePipelineWorkersSmall);
const SIZE_PIPELINE_CPU_SMALL = Number(values.sizePipelineCPUSmall);
const SIZE_PIPELINE_CPU_LARGE = Number(values.sizePipelineCPULarge);
const SIZE_PIPELINE_MEMORY_SMALL = String(values.sizePipelineMemorySmall);
const SIZE_PIPELINE_MEMORY_LARGE = String(values.sizePipelineMemoryLarge);
const NUM_REPOS_PER_PROJECT = DISTRIBUTED
  ? NUM_REPOS / Number(NUM_PROJECTS)
  : NUM_REPOS;
const NUM_PIPELINES_PER_PROJECT = DISTRIBUTED
  ? NUM_PIPELINES / Number(NUM_PROJECTS)
  : NUM_PIPELINES;
const NUM_PROJECTS_WITH_WORK = DISTRIBUTED ? NUM_PROJECTS : 1;
const FIRST_PROJECT_INDEX = DISTRIBUTED ? 1 : 0;

const execCommand = (cmd: string) => {
  console.log('Running...', cmd);

  return execPromise(cmd);
};

const createRepos = async () => {
  // Create 2 repos in perf-project-0 for heavy duty work
  for (const i of range(2)) {
    await execCommand(
      `pachctl create repo perf-repo-${i} --project perf-project-0`,
    );
  }

  // Create many repos for simple work
  for (const i of range(FIRST_PROJECT_INDEX, NUM_PROJECTS_WITH_WORK)) {
    for (const j of range(NUM_REPOS_PER_PROJECT)) {
      await execCommand(
        `pachctl create repo perf-repo-small-${j} --project perf-project-${i}`,
      );
    }
  }
};

const createPipelines = async () => {
  // Create a pipeline to generate a large file
  await execCommand(
    `pachctl create pipeline --jsonnet ./src/pipelines/generate-large-file.jsonnet \\
      --arg name="perf-pipeline-large-files" \\
      --arg inputRepo="perf-repo-0" \\
      --arg sizeFile=${SIZE_FILE_LARGE} \\
      --arg cpuLimit=${SIZE_PIPELINE_CPU_LARGE} \\
      --arg memoryLimit=${SIZE_PIPELINE_MEMORY_LARGE} \\
      --project perf-project-0
    `,
  );

  // Create a pipeline to generate small files
  await execCommand(
    `pachctl create pipeline --jsonnet ./src/pipelines/generate-small-files.jsonnet \\
      --arg name="perf-pipeline-small-files" \\
      --arg inputRepo="perf-repo-1" \\
      --arg numFiles=${NUM_FILES - 1} \\
      --arg sizeFile=${SIZE_FILE_SMALL} \\
      --arg cpuLimit=${SIZE_PIPELINE_CPU_LARGE} \\
      --arg memoryLimit=${SIZE_PIPELINE_MEMORY_LARGE} \\
      --arg workers=${SIZE_PIPELINE_WORKERS_SMALL} \\
      --project perf-project-0
    `,
  );

  // Create a pipeline to log and copy a large file
  await execCommand(
    `pachctl create pipeline --jsonnet ./src/pipelines/log-and-copy-files.jsonnet \\
      --arg name="perf-pipeline-large-0" \\
      --arg inputRepo="perf-pipeline-large-files" \\
      --arg inputFile="0-0.txt" \\
      --arg sizeLogs=${SIZE_LOGS_LARGE} \\
      --arg cpuLimit=${SIZE_PIPELINE_CPU_LARGE} \\
      --arg memoryLimit=${SIZE_PIPELINE_MEMORY_LARGE} \\
      --project perf-project-0
    `,
  );

  // Create an equal number of simple pipelines in other projects
  for (const i of range(FIRST_PROJECT_INDEX, NUM_PROJECTS_WITH_WORK)) {
    for (const j of range(NUM_PIPELINES_PER_PROJECT)) {
      await execCommand(
        `pachctl create pipeline --jsonnet ./src/pipelines/echo-name.jsonnet \\
          --arg name="perf-pipeline-echo-${j}" \\
          --arg inputRepo="perf-repo-small-${j}" \\
          --arg cpuLimit=${SIZE_PIPELINE_CPU_SMALL} \\
          --arg memoryLimit=${SIZE_PIPELINE_MEMORY_SMALL} \\
          --project perf-project-${i}
        `,
      );
    }
  }
};

const createFiles = async () => {
  // Put a file into perf-repo-0 to kickoff the large file pipeline
  await execCommand(
    'echo "0" | pachctl put file perf-repo-0@master:/input-0.txt --project perf-project-0',
  );

  // Put many files into perf-repo-1 to kickoff the small file pipeline
  for (const i of range(NUM_COMMITS)) {
    await execCommand(
      `echo "${i}" | pachctl put file perf-repo-1@master:/input-${i}.txt --project perf-project-0`,
    );
  }

  // Put files in perf-repo-small for each project to kickoff their pipelines
  for (const i of range(FIRST_PROJECT_INDEX, NUM_PROJECTS_WITH_WORK)) {
    for (const j of range(NUM_REPOS_PER_PROJECT)) {
      await execCommand(
        `echo "0" | pachctl put file perf-repo-small-${j}@master:/input-0.txt --project perf-project-${i}`,
      );
    }
  }
};

const createProjects = async () => {
  for (const i of range(NUM_PROJECTS)) {
    await execCommand(`pachctl create project perf-project-${i}`);
  }
};

const assignProjectRoles = async () => {
  for (const i of range(NUM_PROJECTS - NUM_INACCESSIBLE_PROJECTS)) {
    await execCommand(
      `pachctl auth set project perf-project-${i} projectOwner user:kilgore@kilgore.trout`,
    );
  }
  for (const i of range(NUM_INACCESSIBLE_PROJECTS)) {
    await execCommand(
      `pachctl auth set project perf-project-${
        NUM_PROJECTS - 1 - i
      } none user:kilgore@kilgore.trout`,
    );
  }
  if (NUM_INACCESSIBLE_PROJECTS > 0)
    await execCommand(
      `pachctl auth set cluster none user:kilgore@kilgore.trout`,
    );
};

const deleteProjects = async () => {
  for (const i of range(NUM_PROJECTS)) {
    try {
      await execCommand(`yes | pachctl delete project perf-project-${i}`).catch(
        (err) => console.log(err),
      );
    } catch (err) {
      console.error(String(err));
    }
  }
};

const setupCluster = async () => {
  console.log('Setting up cluster with...');
  console.log({
    DISTRIBUTED,
    NUM_PROJECTS,
    NUM_INACCESSIBLE_PROJECTS,
    NUM_REPOS,
    NUM_PIPELINES,
    NUM_COMMITS,
    NUM_FILES,
    SIZE_FILE_SMALL,
    SIZE_FILE_LARGE,
    SIZE_LOGS_LARGE,
    SIZE_PIPELINE_CPU_SMALL,
    SIZE_PIPELINE_CPU_LARGE,
    SIZE_PIPELINE_MEMORY_SMALL,
    SIZE_PIPELINE_MEMORY_LARGE,
    SIZE_PIPELINE_WORKERS_SMALL,
    NUM_REPOS_PER_PROJECT,
    NUM_PIPELINES_PER_PROJECT,
    NUM_PROJECTS_WITH_WORK,
  });

  await createProjects();
  await assignProjectRoles();
  await createRepos();
  await createPipelines();
  await createFiles();
};

const teardownCluster = async () => {
  await deleteProjects();
};

if (MODE === 'setup') {
  setupCluster();
} else if (MODE === 'teardown') {
  teardownCluster();
} else {
  console.log('Doing nothing by default...');
  console.log('Pass --mode=setup or --mode=teardown');
}
