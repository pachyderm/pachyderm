import {
  TaskComponentProps,
  TaskCard,
  MultiSelectModule,
  useMultiSelectModule,
} from '@pachyderm/components';
import React, {useCallback} from 'react';

import {usePutFilesFromUrLsMutation} from '@dash-frontend/generated/hooks';
import useAccount from '@dash-frontend/hooks/useAccount';
import useRecordTutorialProgress from '@dash-frontend/hooks/useRecordTutorialProgress';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const files = {
  'https://i.imgur.com/NUS7JKf.png': {
    name: 'dante.png',
    path: '/dante.png',
  },
  'https://i.imgur.com/1wEn4zc.png': {
    name: 'pippin.jpg',
    path: '/pippin.jpg',
  },
  'https://i.imgur.com/uULNmcY.jpg': {
    name: 'bruce.jpg',
    path: '/bruce.jpg',
  },
  'https://i.imgur.com/9VlTMum.jpg': {
    name: 'kitten.jpg',
    path: '/kitten.jpg',
  },
};

const AddFilesTask: React.FC<TaskComponentProps> = ({
  onCompleted,
  currentTask,
  currentStory,
  index,
  name,
}) => {
  const {projectId} = useUrlState();
  const recordTutorialProgress = useRecordTutorialProgress(
    'image-processing',
    currentStory,
    currentTask,
  );

  const {tutorialId, loading: accountLoading} = useAccount();
  const {register, setDisabled, setUploaded} = useMultiSelectModule({files});

  const [putFilesFromURLsMutation, {loading, error}] =
    usePutFilesFromUrLsMutation({
      onCompleted: () => {
        setUploaded();
        onCompleted();
        recordTutorialProgress();
      },
      onError: () => setDisabled(false),
    });

  const action = useCallback(() => {
    if (!loading) {
      setDisabled(true);
      putFilesFromURLsMutation({
        variables: {
          args: {
            files: Object.keys(register.files)
              .filter((url) => register.files[url].selected)
              .map((url) => ({url, path: register.files[url].path})),
            branch: 'master',
            repo: `images_${tutorialId}`,
            projectId,
          },
        },
      });
    }
  }, [
    loading,
    setDisabled,
    putFilesFromURLsMutation,
    register.files,
    tutorialId,
    projectId,
  ]);

  return (
    <TaskCard
      task={name}
      index={index}
      action={action}
      error={error?.message}
      disabled={accountLoading}
      currentTask={currentTask}
      actionText="Add these images"
      taskInfoTitle="Adding images to see the montage change"
      taskInfo={
        <p>
          Select one or more images and click &quot;add&quot; get them added to
          the images input repo. The images get processed automatically by
          Pachyderm&apos;s data-driven pipelines, and when the job is finished,
          the montage is updated with the new images.
        </p>
      }
    >
      <MultiSelectModule
        repo={`images_${tutorialId}`}
        type="image"
        {...register}
      />
    </TaskCard>
  );
};

export default AddFilesTask;
