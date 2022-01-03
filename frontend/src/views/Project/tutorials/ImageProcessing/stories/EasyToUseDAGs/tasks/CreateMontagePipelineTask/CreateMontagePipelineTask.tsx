import {
  TaskComponentProps,
  TaskCard,
  LoadingDots,
  ConfigurationUploadModule,
} from '@pachyderm/components';
import React from 'react';

import useCreatePipeline from '@dash-frontend/hooks/useCreatePipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';

// just used for display, make sure to update the useCreatePipeline args if this is updated
const PIPELINE_YAML = `
---
pipeline:
  name: montage
description: A pipeline that combines images from the \`images\` and \`edges\` repositories into a montage.
input:
  cross:
    - pfs:
        glob: /
        repo: images
    - pfs:
        glob: /
        repo: edges
transform:
  cmd:
    - sh
  image: dpokidov/imagemagick:7.0.10-58
  stdin:
    - montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png

`;

const CreateMontagePipelineTask: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  index,
  name,
}) => {
  const {projectId} = useUrlState();

  const {createPipeline, status} = useCreatePipeline(
    {
      name: 'montage',
      description:
        'A pipeline that combines images from the `images` and `edges` repositories into a montage.',
      transform: {
        image: 'dpokidov/imagemagick:7.0.10-58',
        cmdList: ['sh'],
        stdinList: [
          'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
        ],
      },
      crossList: [
        {name: 'images', repo: {name: 'images'}, glob: '/'},
        {name: 'edges', repo: {name: 'edges'}, glob: '/'},
      ],
      projectId: projectId,
    },
    onCompleted,
  );

  const file = {
    name: 'montage.yaml',
    path: 'https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/opencv/montage.yaml',
    contents: PIPELINE_YAML,
  };

  return (
    <TaskCard
      task={name}
      index={index}
      action={createPipeline}
      currentTask={currentTask}
      actionText="Create the montage pipeline"
      taskInfoTitle="Create the montage pipeline"
      taskInfo={
        <p>
          The montage pipeline will use both the images repo and the edges
          pipeline as its input.
        </p>
      }
    >
      {status.loading ? <LoadingDots /> : null}
      <ConfigurationUploadModule file={file} />
    </TaskCard>
  );
};

export default CreateMontagePipelineTask;
