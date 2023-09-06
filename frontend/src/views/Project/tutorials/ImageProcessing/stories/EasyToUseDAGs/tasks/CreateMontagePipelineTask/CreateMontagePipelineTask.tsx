import React from 'react';

import useAccount from '@dash-frontend/hooks/useAccount';
import useCreatePipeline from '@dash-frontend/hooks/useCreatePipeline';
import useRecordTutorialProgress from '@dash-frontend/hooks/useRecordTutorialProgress';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  TaskComponentProps,
  TaskCard,
  LoadingDots,
  ConfigurationUploadModule,
} from '@pachyderm/components';

// just used for display, make sure to update the useCreatePipeline args if this is updated
const PIPELINE_JSON = `{
  pipeline: {
    name: 'montage',
  },
  description:
    'A pipeline that combines images from the images and edges repositories into a montage.',
  input: {
    cross: [
      {
        pfs: {
          glob: '/',
          repo: 'images',
        },
      },
      {
        pfs: {
          glob: '/',
          repo: 'edges',
        },
      },
    ],
  },
  transform: {
    cmd: ['sh'],
    image: 'dpokidov/imagemagick:7.1.0-23',
    stdin: [
      'montage -shadow -background SkyBlue -geometry 300x300+2+2 $(find /pfs -type f | sort) /pfs/out/montage.png',
    ],
  },
}`;

const CreateMontagePipelineTask: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  currentStory,
  index,
  name,
}) => {
  const {projectId} = useUrlState();
  const {tutorialId, loading: accountLoading} = useAccount();
  const recordTutorialProgress = useRecordTutorialProgress(
    'image-processing',
    currentStory,
    currentTask,
  );

  const {createPipeline, status} = useCreatePipeline(
    {
      name: `montage_${tutorialId}`,
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
        {
          name: `images_montage_${tutorialId}`,
          repo: {name: `images_${tutorialId}`},
          glob: '/',
        },
        {
          name: `edges_${tutorialId}`,
          repo: {name: `edges_${tutorialId}`},
          glob: '/',
        },
      ],
      projectId: projectId,
    },
    () => {
      onCompleted();
      recordTutorialProgress();
    },
  );

  const file = {
    name: 'montage.json',
    path: 'https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/opencv/montage.pipeline.json',
  };

  return (
    <TaskCard
      task={name}
      index={index}
      action={createPipeline}
      error={status.error?.message}
      currentTask={currentTask}
      actionText="Create the montage pipeline"
      taskInfoTitle="Create the montage pipeline"
      taskInfo={
        <p>
          The montage pipeline will use both the images repo and the edges
          pipeline as its input.
        </p>
      }
      disabled={accountLoading}
    >
      {status.loading ? <LoadingDots /> : null}
      <ConfigurationUploadModule fileMeta={file} fileContents={PIPELINE_JSON} />
    </TaskCard>
  );
};

export default CreateMontagePipelineTask;
