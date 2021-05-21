import {
  BuildSpec,
  ChunkSpec,
  CronInput,
  Egress,
  GitInput,
  GPUSpec,
  Input,
  PipelineJobState,
  ParallelismSpec,
  PFSInput,
  Pipeline,
  PipelineInfo,
  PipelineInfos,
  PipelineState,
  ResourceSpec,
  SchedulingSpec,
  SecretMount,
  Service,
  Spout,
  TFJob,
  Transform,
  PipelineJob,
  PipelineJobInfo,
} from '@pachyderm/proto/pb/pps/pps_pb';

import {
  commitFromObject,
  CommitObject,
  triggerFromObject,
  TriggerObject,
} from './pfs';
import {
  durationFromObject,
  DurationObject,
  timestampFromObject,
  TimestampObject,
} from './protobuf';

export type PipelineObject = {
  name: Pipeline.AsObject['name'];
};

export type SecretMountObject = {
  name: SecretMount.AsObject['name'];
  key: SecretMount.AsObject['key'];
  mountPath: SecretMount.AsObject['mountPath'];
  envVar: SecretMount.AsObject['envVar'];
};

export type BuildSpecObject = {
  path: BuildSpec.AsObject['path'];
  language: BuildSpec.AsObject['language'];
  image: BuildSpec.AsObject['image'];
};

export type TransformObject = {
  image: Transform.AsObject['image'];
  cmdList: Transform.AsObject['cmdList'];
  errCmdList?: Transform.AsObject['errCmdList'];
  //TODO: Proto Map does not have a setter
  // envMap: Array<[string, string]>;
  secretsList?: SecretMountObject[];
  imagePullSecretsList?: Transform.AsObject['imagePullSecretsList'];
  stdinList?: Transform.AsObject['stdinList'];
  errStdinList?: Transform.AsObject['errStdinList'];
  acceptReturnCodeList?: Transform.AsObject['acceptReturnCodeList'];
  debug?: Transform.AsObject['debug'];
  user?: Transform.AsObject['user'];
  workingDir?: Transform.AsObject['workingDir'];
  dockerfile?: Transform.AsObject['dockerfile'];
  build?: BuildSpecObject;
};

export type TFJobObject = {
  tfJob: TFJob.AsObject['tfJob'];
};

export type ParallelismSpecObject = {
  constant: ParallelismSpec.AsObject['constant'];
  coefficient: ParallelismSpec.AsObject['coefficient'];
};

export type EgressObject = {
  url: Egress.AsObject['url'];
};

export type GPUSpecObject = {
  type: GPUSpec.AsObject['type'];
  number: GPUSpec.AsObject['number'];
};

export type ResourceSpecObject = {
  cpu: ResourceSpec.AsObject['cpu'];
  memory: ResourceSpec.AsObject['memory'];
  gpu?: GPUSpecObject;
  disk: ResourceSpec.AsObject['disk'];
};

export type PFSInputObject = {
  name: PFSInput.AsObject['name'];
  repo: PFSInput.AsObject['repo'];
  branch: PFSInput.AsObject['branch'];
  commit?: PFSInput.AsObject['commit'];
  glob?: PFSInput.AsObject['glob'];
  joinOn?: PFSInput.AsObject['joinOn'];
  outerJoin?: PFSInput.AsObject['outerJoin'];
  groupBy?: PFSInput.AsObject['groupBy'];
  lazy?: PFSInput.AsObject['lazy'];
  emptyFiles?: PFSInput.AsObject['emptyFiles'];
  s3?: PFSInput.AsObject['s3'];
  trigger?: TriggerObject;
};

export type CronInputObject = {
  name: CronInput.AsObject['name'];
  repo: CronInput.AsObject['repo'];
  commit: CronInput.AsObject['commit'];
  spec: CronInput.AsObject['spec'];
  overwrite: CronInput.AsObject['overwrite'];
  start?: TimestampObject;
};

export type GitInputObject = {
  name: GitInput.AsObject['name'];
  url: GitInput.AsObject['url'];
  branch: GitInput.AsObject['branch'];
  commit: GitInput.AsObject['commit'];
};

export type InputObject = {
  pfs?: PFSInputObject;
  joinList?: InputObject[];
  groupList?: InputObject[];
  crossList?: InputObject[];
  unionList?: InputObject[];
  cron?: CronInputObject;
  git?: GitInputObject;
};
export type ServiceObject = {
  internalPort: Service.AsObject['internalPort'];
  externalPort: Service.AsObject['externalPort'];
  ip: Service.AsObject['ip'];
  type: Service.AsObject['type'];
};

export type SpoutObject = {
  service?: ServiceObject;
};

export type ChunkSpecObject = {
  number: ChunkSpec.AsObject['number'];
  sizeBytes: ChunkSpec.AsObject['sizeBytes'];
};
export type SchedulingSpecObject = {
  //TODO: Proto Map does not have a setter
  // nodeSelectorMap: jspb.Map<string, string>;
  priorityClassName: SchedulingSpec.AsObject['priorityClassName'];
};

export type PipelineInfoObject = {
  pipeline?: PipelineObject;
  version?: PipelineInfo.AsObject['version'];
  transform?: TransformObject;
  tfJob?: TFJobObject;
  parallelismSpec?: ParallelismSpecObject;
  egress?: EgressObject;
  createdAt?: TimestampObject;
  state?: PipelineState;
  stopped?: PipelineInfo.AsObject['stopped'];
  recentError?: PipelineInfo.AsObject['recentError'];
  workersRequested?: PipelineInfo.AsObject['workersRequested'];
  workersAvailable?: PipelineInfo.AsObject['workersAvailable'];
  //TODO: Proto Map does not have a setter
  // jobCountsMap: jspb.Map<number, number>;
  lastJobState?: PipelineJobState;
  outputBranch?: PipelineInfo.AsObject['outputBranch'];
  resourceRequests?: ResourceSpecObject;
  resourceLimits?: ResourceSpecObject;
  sidecarResourceLimits?: ResourceSpecObject;
  input?: InputObject;
  description?: PipelineInfo.AsObject['description'];
  cacheSize?: PipelineInfo.AsObject['cacheSize'];
  enableStats?: PipelineInfo.AsObject['enableStats'];
  salt?: PipelineInfo.AsObject['salt'];
  reason?: PipelineInfo.AsObject['reason'];
  maxQueueSize?: PipelineInfo.AsObject['maxQueueSize'];
  service?: ServiceObject;
  spout?: SpoutObject;
  chunkSpec?: ChunkSpecObject;
  datumTimeout?: DurationObject;
  jobTimeout?: DurationObject;
  githookUrl?: PipelineInfo.AsObject['githookUrl'];
  specCommit?: CommitObject;
  standby?: PipelineInfo.AsObject['standby'];
  datumTries?: PipelineInfo.AsObject['datumTries'];
  schedulingSpec?: SchedulingSpecObject;
  podSpec?: PipelineInfo.AsObject['podSpec'];
  podPatch?: PipelineInfo.AsObject['podPatch'];
  s3Out?: PipelineInfo.AsObject['s3Out'];
  //TODO: Proto Map does not have a setter
  // metadata?: Metadata.AsObject,
};

export type PipelineInfosObject = {
  pipelineInfoList: PipelineInfoObject[];
};

export type PipelineJobObject = {
  id: PipelineJob.AsObject['id'];
};

export type PipelineJobInfoObject = {
  pipelineJob: Pick<PipelineJob.AsObject, 'id'>;
  createdAt: PipelineJobInfo.AsObject['started'];
  state: PipelineJobState;
  pipeline: {
    name: Pipeline.AsObject['name'];
  };
};

export const pipelineFromObject = ({name}: PipelineObject) => {
  const pipeline = new Pipeline();
  pipeline.setName(name);

  return pipeline;
};

export const secretMountFromObject = ({
  name,
  key,
  mountPath,
  envVar,
}: SecretMountObject) => {
  const secretMount = new SecretMount();
  secretMount.setName(name);
  secretMount.setKey(key);
  secretMount.setMountPath(mountPath);
  secretMount.setEnvVar(envVar);

  return secretMount;
};

export const buildSpecFromObject = ({
  path,
  language,
  image,
}: BuildSpecObject) => {
  const buildSpec = new BuildSpec();
  buildSpec.setPath(path);
  buildSpec.setLanguage(language);
  buildSpec.setImage(image);

  return buildSpec;
};

export const transformFromObject = ({
  image,
  cmdList,
  errCmdList = [],
  secretsList = [],
  imagePullSecretsList = [],
  stdinList = [],
  errStdinList = [],
  acceptReturnCodeList = [],
  debug = false,
  user = '',
  workingDir = '',
  dockerfile = '',
  build,
}: TransformObject) => {
  const transform = new Transform();
  transform.setImage(image);
  transform.setCmdList(cmdList);
  transform.setErrCmdList(errCmdList);

  if (secretsList) {
    const transformedSecrets = secretsList.map((secret) => {
      return secretMountFromObject(secret);
    });
    transform.setSecretsList(transformedSecrets);
  }

  transform.setImagePullSecretsList(imagePullSecretsList);
  transform.setStdinList(stdinList);
  transform.setErrStdinList(errStdinList);
  transform.setAcceptReturnCodeList(acceptReturnCodeList);
  transform.setDebug(debug);
  transform.setUser(user);
  transform.setWorkingDir(workingDir);
  transform.setDockerfile(dockerfile);
  if (build) {
    transform.setBuild(buildSpecFromObject(build));
  }

  return transform;
};

export const tfJobFromObject = ({tfJob}: TFJobObject) => {
  const tfjJob = new TFJob();
  tfjJob.setTfJob(tfJob);

  return tfjJob;
};

export const parallelismSpecFromObject = ({
  constant,
  coefficient,
}: ParallelismSpecObject) => {
  const parallelismSpec = new ParallelismSpec();
  parallelismSpec.setConstant(constant);
  parallelismSpec.setCoefficient(coefficient);

  return parallelismSpec;
};

export const egressFromObject = ({url}: EgressObject) => {
  const egress = new Egress();
  egress.setUrl(url);

  return egress;
};

export const gpuSpecFromObject = ({type, number}: GPUSpecObject) => {
  const gpuSpec = new GPUSpec();
  gpuSpec.setType(type);
  gpuSpec.setNumber(number);

  return gpuSpec;
};

export const resourceSpecFromObject = ({
  cpu,
  memory,
  gpu,
  disk,
}: ResourceSpecObject) => {
  const resourceSpec = new ResourceSpec();
  resourceSpec.setCpu(cpu);
  resourceSpec.setMemory(memory);
  if (gpu) {
    resourceSpec.setGpu(gpuSpecFromObject(gpu));
  }
  resourceSpec.setDisk(disk);
  return resourceSpec;
};

export const pfsInputFromObject = ({
  name,
  repo,
  branch,
  commit = '',
  glob = '',
  joinOn = '',
  outerJoin = false,
  groupBy = '',
  lazy = false,
  emptyFiles = false,
  s3 = false,
  trigger,
}: PFSInputObject) => {
  const pfsInput = new PFSInput();
  pfsInput.setName(name);
  pfsInput.setRepo(repo);
  pfsInput.setBranch(branch);
  pfsInput.setCommit(commit);
  pfsInput.setGlob(glob);
  pfsInput.setJoinOn(joinOn);
  pfsInput.setOuterJoin(outerJoin);
  pfsInput.setGroupBy(groupBy);
  pfsInput.setLazy(lazy);
  pfsInput.setEmptyFiles(emptyFiles);
  pfsInput.setS3(s3);
  if (trigger) {
    pfsInput.setTrigger(triggerFromObject(trigger));
  }

  return pfsInput;
};

export const cronInputFromObject = ({
  name,
  repo,
  commit,
  spec,
  overwrite,
  start,
}: CronInputObject) => {
  const cronInput = new CronInput();
  cronInput.setName(name);
  cronInput.setRepo(repo);
  cronInput.setCommit(commit);
  cronInput.setSpec(spec);
  cronInput.setOverwrite(overwrite);

  if (start) {
    cronInput.setStart(timestampFromObject(start));
  }
  return cronInput;
};

export const gitInputFromObject = ({
  name,
  url,
  branch,
  commit,
}: GitInputObject) => {
  const gitInput = new GitInput();
  gitInput.setName(name);
  gitInput.setUrl(url);
  gitInput.setBranch(branch);
  gitInput.setCommit(commit);

  return gitInput;
};

export const inputFromObject = ({
  pfs,
  joinList = [],
  groupList = [],
  crossList = [],
  unionList = [],
  cron,
  git,
}: InputObject) => {
  const input = new Input();

  if (pfs) {
    input.setPfs(pfsInputFromObject(pfs));
  }

  if (joinList) {
    const transformedJoinlist = joinList.map((item) => {
      return inputFromObject(item);
    });
    input.setJoinList(transformedJoinlist);
  }

  if (groupList) {
    const transformedGroupList = groupList.map((item) => {
      return inputFromObject(item);
    });
    input.setGroupList(transformedGroupList);
  }

  if (crossList) {
    const transformedCrosslist = crossList.map((item) => {
      return inputFromObject(item);
    });
    input.setCrossList(transformedCrosslist);
  }

  if (unionList) {
    const transformedUnionlist = unionList.map((item) => {
      return inputFromObject(item);
    });
    input.setUnionList(transformedUnionlist);
  }

  if (cron) {
    input.setCron(cronInputFromObject(cron));
  }
  if (git) {
    input.setGit(gitInputFromObject(git));
  }

  return input;
};

export const serviceFromObject = ({
  internalPort,
  externalPort,
  ip,
  type,
}: ServiceObject) => {
  const service = new Service();
  service.setInternalPort(internalPort);
  service.setExternalPort(externalPort);
  service.setIp(ip);
  service.setType(type);

  return service;
};

export const spoutFromObject = ({service}: SpoutObject) => {
  const spout = new Spout();
  if (service) {
    spout.setService(serviceFromObject(service));
  }

  return spout;
};

export const chunkSpecFromObject = ({number, sizeBytes}: ChunkSpecObject) => {
  const chunkSpec = new ChunkSpec();
  chunkSpec.setNumber(number);
  chunkSpec.setSizeBytes(sizeBytes);

  return chunkSpec;
};

export const schedulingSpecFromObject = ({
  priorityClassName,
}: SchedulingSpecObject) => {
  const schedulingSpec = new SchedulingSpec();
  schedulingSpec.setPriorityClassName(priorityClassName);

  return schedulingSpec;
};

export const pipelineInfoFromObject = ({
  pipeline,
  version = 1,
  transform,
  tfJob,
  parallelismSpec,
  egress,
  createdAt,
  state = 0,
  stopped = false,
  recentError = '',
  workersRequested = 0,
  workersAvailable = 0,
  lastJobState = 0,
  outputBranch = 'master',
  resourceRequests,
  resourceLimits,
  sidecarResourceLimits,
  input,
  description = '',
  cacheSize = '',
  enableStats = false,
  salt = '',
  reason = '',
  maxQueueSize = 1,
  service,
  spout,
  chunkSpec,
  datumTimeout,
  jobTimeout,
  githookUrl = '',
  specCommit,
  standby = false,
  datumTries = 0,
  podSpec = '',
  podPatch = '',
  s3Out = false,
}: PipelineInfoObject) => {
  const pipelineInfo = new PipelineInfo();

  if (pipeline) {
    pipelineInfo.setPipeline(pipelineFromObject(pipeline));
  }
  pipelineInfo.setVersion(version);
  if (transform) {
    pipelineInfo.setTransform(transformFromObject(transform));
  }
  if (tfJob) {
    pipelineInfo.setTfJob(tfJobFromObject(tfJob));
  }
  if (parallelismSpec) {
    pipelineInfo.setParallelismSpec(parallelismSpecFromObject(parallelismSpec));
  }
  if (egress) {
    pipelineInfo.setEgress(egressFromObject(egress));
  }
  if (createdAt) {
    pipelineInfo.setCreatedAt(timestampFromObject(createdAt));
  }

  pipelineInfo.setState(state);
  pipelineInfo.setStopped(stopped);
  pipelineInfo.setRecentError(recentError);
  pipelineInfo.setWorkersRequested(workersRequested);
  pipelineInfo.setWorkersAvailable(workersAvailable);
  pipelineInfo.setLastJobState(lastJobState);
  pipelineInfo.setOutputBranch(outputBranch);

  if (resourceRequests) {
    pipelineInfo.setResourceRequests(resourceSpecFromObject(resourceRequests));
  }
  if (resourceLimits) {
    pipelineInfo.setResourceLimits(resourceSpecFromObject(resourceLimits));
  }
  if (sidecarResourceLimits) {
    pipelineInfo.setSidecarResourceLimits(
      resourceSpecFromObject(sidecarResourceLimits),
    );
  }
  if (input) {
    pipelineInfo.setInput(inputFromObject(input));
  }
  if (service) {
    pipelineInfo.setService(serviceFromObject(service));
  }
  if (spout) {
    pipelineInfo.setSpout(spoutFromObject(spout));
  }
  if (chunkSpec) {
    pipelineInfo.setChunkSpec(chunkSpecFromObject(chunkSpec));
  }
  if (datumTimeout) {
    pipelineInfo.setDatumTimeout(durationFromObject(datumTimeout));
  }
  if (jobTimeout) {
    pipelineInfo.setJobTimeout(durationFromObject(jobTimeout));
  }
  if (specCommit) {
    pipelineInfo.setSpecCommit(commitFromObject(specCommit));
  }

  pipelineInfo.setDescription(description);
  pipelineInfo.setCacheSize(cacheSize);
  pipelineInfo.setEnableStats(enableStats);
  pipelineInfo.setSalt(salt);
  pipelineInfo.setReason(reason);
  pipelineInfo.setMaxQueueSize(maxQueueSize);
  pipelineInfo.setGithookUrl(githookUrl);
  pipelineInfo.setStandby(standby);
  pipelineInfo.setDatumTries(datumTries);
  pipelineInfo.setPodSpec(podSpec);
  pipelineInfo.setPodPatch(podPatch);
  pipelineInfo.setS3Out(s3Out);

  return pipelineInfo;
};

export const pipelineInfosFromObject = ({
  pipelineInfoList,
}: PipelineInfosObject) => {
  const pipelineInfos = new PipelineInfos();
  if (pipelineInfoList) {
    const transformedPipelines = pipelineInfoList.map((item) => {
      return pipelineInfoFromObject(item);
    });
    pipelineInfos.setPipelineInfoList(transformedPipelines);
  }

  return pipelineInfos;
};

export const pipelineJobFromObject = ({id}: PipelineJobObject) => {
  const job = new PipelineJob();
  job.setId(id);

  return job;
};

export const pipelineJobInfoFromObject = ({
  pipelineJob: {id},
  createdAt,
  state,
  pipeline: {name},
}: PipelineJobInfoObject) => {
  const jobInfo = new PipelineJobInfo()
    .setState(state)
    .setStarted(
      timestampFromObject({seconds: createdAt?.seconds || 0, nanos: 0}),
    )
    .setPipelineJob(new PipelineJob().setId(id))
    .setPipeline(new Pipeline().setName(name));

  return jobInfo;
};
