import {
  TaskComponentProps,
  TaskCard,
  LoadingDots,
  ConfigurationUploadModule,
} from '@pachyderm/components';
import React from 'react';

import useCreatePipeline from '@dash-frontend/hooks/useCreatePipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const PIPELINE_YAML = `
---
pipeline:
  name: edges
input:
  pfs:
    glob: /*
    repo: images
transform:
  cmd:
  - python3
  - "/edges.py"
  image: pachyderm/opencv@sha256:38e947bccc9c8e1034026033c6fb1e605e35929caf4f7fe11115cf5a1e84859a
`;

const CreatePipelineTask: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
}) => {
  const {projectId} = useUrlState();

  const {createPipeline, status} = useCreatePipeline(
    {
      name: 'edges',
      transform: {
        image: 'pachyderm/opencv',
        cmdList: ['python3', '/edges.py'],
      },
      pfs: {name: 'images', repo: {name: 'images'}, glob: '/*'},
      projectId: projectId,
    },
    onCompleted,
  );

  const file = {
    name: 'pipeline.yaml',
    path: 'https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/opencv/edges.yaml',
    contents: PIPELINE_YAML,
  };

  return (
    <TaskCard
      task={name}
      index={index}
      action={createPipeline}
      currentTask={currentTask}
      actionText="Create the edges pipeline"
      taskInfoTitle="Create the edges pipeline"
      taskInfo={
        <>
          <p>
            The edges pipeline will use the images repo that you created as its
            input. If the repo already exists, you don&apos;t need to create it.
            More than one pipeline can use a repo as input, and a pipeline can
            have more than one repo as input.
          </p>
          <p>The definition of this pipeline looks like this, in YAML</p>
        </>
      }
    >
      {status.loading ? <LoadingDots /> : null}
      <ConfigurationUploadModule file={file} />
    </TaskCard>
  );
};

export default CreatePipelineTask;
