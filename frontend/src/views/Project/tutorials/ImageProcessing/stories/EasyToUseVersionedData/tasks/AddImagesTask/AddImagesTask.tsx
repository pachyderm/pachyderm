import React, {useCallback} from 'react';

import {usePutFilesFromUrLsMutation} from '@dash-frontend/generated/hooks';
import useAccount from '@dash-frontend/hooks/useAccount';
import useRecordTutorialProgress from '@dash-frontend/hooks/useRecordTutorialProgress';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  MultiSelectModule,
  TaskCard,
  TaskComponentProps,
  useMultiSelectModule,
} from '@pachyderm/components';

const files = {
  'https://i.imgur.com/FOO9q43.jpg': {
    name: 'birthday-cake.jpg',
    path: '/birthday-cake.jpg',
  },
  'https://i.imgur.com/8rfjLgG.jpg': {
    name: 'sutro-tower.jpg',
    path: '/sutro-tower.jpg',
  },
  'https://i.imgur.com/SpYheUO.jpg': {
    name: 'wine.jpg',
    path: '/wine.jpg',
  },
  'https://i.imgur.com/WzHCuVx.jpg': {
    name: 'puppy.jpg',
    path: '/puppy.jpg',
  },
};

const AddImagesTask: React.FC<TaskComponentProps> = ({
  currentTask,
  currentStory,
  onCompleted,
  index,
  name,
}) => {
  const {register, setDisabled, setUploaded} = useMultiSelectModule({files});
  const {projectId} = useUrlState();
  const recordTutorialProgress = useRecordTutorialProgress(
    'image-processing',
    currentStory,
    currentTask,
  );
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
      taskInfoTitle="Adding images for processing"
      taskInfo={
        'Select one or more images and click "add" to get them added to the images input repo. Since we already added the edges pipeline, you\'ll see them get processed automatically.'
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
