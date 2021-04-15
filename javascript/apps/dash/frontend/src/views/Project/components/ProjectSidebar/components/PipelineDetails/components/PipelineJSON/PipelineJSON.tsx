import {SkeletonBodyText} from '@pachyderm/components';
import React from 'react';

import JSONBlock from '@dash-frontend/components/JSONBlock';
import PipelineInput from '@dash-frontend/components/PipelineInput';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import Description from '../Description';

import styles from './PipelineJSON.module.css';

const Skeleton = () => {
  return (
    <dl>
      <Description term="Inputs">
        <SkeletonBodyText lines={5} />
      </Description>

      <Description term="Transform">
        <SkeletonBodyText lines={3} />
      </Description>
    </dl>
  );
};

const PipelineJSON = () => {
  const {projectId} = useUrlState();
  const {pipeline, loading} = useCurrentPipeline();

  if (loading || !pipeline) {
    return <Skeleton />;
  }

  return (
    <dl>
      <Description term="Inputs">
        <PipelineInput
          inputString={pipeline.inputString}
          projectId={projectId}
          className={styles.input}
        />
      </Description>

      <Description term="Transform">
        {pipeline.transform ? (
          <JSONBlock>
            {JSON.stringify(
              {
                image: pipeline.transform.image,
                cmd: pipeline.transform.cmdList,
              },
              null,
              2,
            )}
          </JSONBlock>
        ) : (
          'N/A'
        )}
      </Description>
    </dl>
  );
};

export default PipelineJSON;
