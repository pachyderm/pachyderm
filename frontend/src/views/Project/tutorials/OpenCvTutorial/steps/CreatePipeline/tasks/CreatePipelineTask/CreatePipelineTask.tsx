import {TaskComponentProps, TaskCard, LoadingDots} from '@pachyderm/components';
import React from 'react';

import useCreatePipeline from '@dash-frontend/hooks/useCreatePipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import styles from './CreatePipelineTask.module.css';

const PIPELINE_JSON = `
{
  "pipeline": {
    "name": "edges"
  },
  "description": "A pipeline that performs image edge detection by using the OpenCV library.",
  "transform": {
    "cmd": [ "python3", "/edges.py" ],
    "image": "pachyderm/opencv"
  },
  "input": {
    "pfs": {
      "repo": "images",
      "glob": "/*"
    }
  }
}
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
      image: 'pachyderm/opencv',
      cmdList: ['python3', '/edges.py'],
      pfs: {name: 'images', repo: {name: 'images'}, glob: '/*'},
      projectId: projectId,
    },
    onCompleted,
  );

  return (
    <TaskCard
      index={index}
      name={name}
      action={createPipeline}
      currentTask={currentTask}
      actionText={'Create pipeline spec'}
    >
      {status.loading ? <LoadingDots /> : null}
      {currentTask <= index ? (
        <div>
          <div className={styles.splitCode}>
            <div className={styles.json}>
              <pre>{PIPELINE_JSON}</pre>
            </div>
            <div className={styles.console}>
              <pre>
                {`> pachctl create pipeline -f https://raw.githubusercontent.com/pachyderm/pachyderm/1.13.x/examples/opencv/edges.json`}
              </pre>
            </div>
          </div>
        </div>
      ) : null}
    </TaskCard>
  );
};

export default CreatePipelineTask;
