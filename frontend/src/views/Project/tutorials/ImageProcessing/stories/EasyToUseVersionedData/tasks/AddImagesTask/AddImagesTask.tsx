import {
  MultiSelectModule,
  TaskCard,
  TaskComponentProps,
  useMultiSelectModule,
} from '@pachyderm/components';
import React, {useCallback} from 'react';

import {usePutFilesFromUrLsMutation} from '@dash-frontend/generated/hooks';
import useUrlState from '@dash-frontend/hooks/useUrlState';

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
  onCompleted,
  index,
  name,
}) => {
  const {register, setDisabled, setUploaded} = useMultiSelectModule({files});
  const {projectId} = useUrlState();

  const [putFilesFromURLsMutation, {loading}] = usePutFilesFromUrLsMutation({
    onCompleted: () => {
      setUploaded();
      onCompleted();
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
      variables: {args: {files, branch: 'master', repo: 'images', projectId}},
    });
  }, [register.files, putFilesFromURLsMutation, projectId]);

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
    >
      <MultiSelectModule type="image" repo="images" {...register} />
    </TaskCard>
  );
};

export default AddImagesTask;
