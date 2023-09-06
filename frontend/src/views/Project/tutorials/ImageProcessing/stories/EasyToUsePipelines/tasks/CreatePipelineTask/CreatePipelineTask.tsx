import React from 'react';

import CodePreview from '@dash-frontend/components/CodePreview';
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

const PIPELINE_JSON = `{
  "pipeline": {
    "name": "edges"
  },
  "input": {
    "pfs": {
      "glob": "/*",
      "repo": "images"
    }
  },
  "transform": {
    "cmd": [ "python3", "/edges.py" ],
    "image": "pachyderm/opencv:1.0"
  }
}`;

const CreatePipelineTask: React.FC<TaskComponentProps> = ({
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
      name: `edges_${tutorialId}`,
      transform: {
        image: 'pachyderm/console-opencv:latest',
        cmdList: ['python3', '/edges.py', tutorialId],
      },
      pfs: {
        name: `images_${tutorialId}`,
        repo: {name: `images_${tutorialId}`},
        glob: '/*',
      },
      projectId: projectId,
    },
    () => {
      onCompleted();
      recordTutorialProgress();
    },
  );

  const file = {
    name: 'pipeline.json',
    path: 'https://raw.githubusercontent.com/pachyderm/pachyderm/master/examples/opencv/edges.pipeline.json',
  };

  return (
    <TaskCard
      task={name}
      index={index}
      action={createPipeline}
      error={status.error?.message}
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
          <p>The definition of this pipeline looks like this, in JSON</p>
        </>
      }
      disabled={accountLoading}
    >
      {status.loading ? <LoadingDots /> : null}
      <ConfigurationUploadModule fileMeta={file}>
        <CodePreview source={PIPELINE_JSON} language="json" />
      </ConfigurationUploadModule>
    </TaskCard>
  );
};

export default CreatePipelineTask;
