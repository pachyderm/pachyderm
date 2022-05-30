import {
  MultiSelectModule,
  TaskCard,
  TaskComponentProps,
  useMultiSelectModule,
} from '@pachyderm/components';
import React, {useCallback} from 'react';

import {usePutFilesFromUrLsMutation} from '@dash-frontend/generated/hooks';
import useAccount from '@dash-frontend/hooks/useAccount';
import useRecordTutorialProgress from '@dash-frontend/hooks/useRecordTutorialProgress';
import useUrlState from '@dash-frontend/hooks/useUrlState';

const files = {
  'https://imgur.com/Togu2RY.jpg': {
    name: 'pooh.jpg',
    path: '/pooh.jpg',
  },
  'https://i.imgur.com/46Q8nDz.jpg': {
    name: 'statue-of-liberty.jpg',
    path: '/statue-of-liberty.jpg',
  },
  'https://i.imgur.com/g2QnNqa.jpg': {
    name: 'kitten.jpg',
    path: '/kitten.jpg',
  },
  'https://i.imgur.com/8MN9Kg0.jpg': {
    name: 'at-at.jpg',
    path: '/at-at.jpg',
  },
};

const AddImagesTask: React.FC<TaskComponentProps> = ({
  currentTask,
  currentStory,
  onCompleted,
  index,
  name,
}) => {
  const {projectId} = useUrlState();
  const recordTutorialProgress = useRecordTutorialProgress(
    'image-processing',
    currentStory,
    currentTask,
  );
  const {register, setDisabled, setUploaded} = useMultiSelectModule({files});
  const {tutorialId, loading: accountLoading} = useAccount();

  const [putFilesFromURLsMutation, {loading, error}] =
    usePutFilesFromUrLsMutation({
      onCompleted: () => {
        setUploaded();
        onCompleted();
        recordTutorialProgress();
      },
      onError: () => setDisabled(false),
    });

  const addFiles = useCallback(() => {
    const files = Object.keys(register.files)
      .filter(
        (key) => register.files[key].selected && !register.files[key].uploaded,
      )
      .map((file) => ({url: file, path: register.files[file].path}));
    putFilesFromURLsMutation({
      variables: {
        args: {
          files,
          branch: 'master',
          repo: `images_${tutorialId}`,
          projectId,
        },
      },
    });
  }, [register.files, putFilesFromURLsMutation, projectId, tutorialId]);

  const action = useCallback(() => {
    if (!loading) {
      setDisabled(true);
      addFiles();
    }
  }, [addFiles, loading, setDisabled]);

  return (
    <TaskCard
      task={name}
      index={index}
      currentTask={currentTask}
      actionText={'Add these images'}
      taskInfoTitle="Adding images to see the montage change"
      taskInfo={
        'Select one or more images and click "add" to get them added to the images input repo. Only the new images are processed by edges, since Pachyderm\'s data-driven pipelines and data lineage are keeping track of everything.'
      }
      action={action}
      error={error?.message}
      disabled={accountLoading}
    >
      <MultiSelectModule
        type="image"
        repo={`images_${tutorialId}`}
        {...register}
      />
    </TaskCard>
  );
};

export default AddImagesTask;
