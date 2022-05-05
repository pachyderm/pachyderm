import {SkeletonBodyText} from '@pachyderm/components';
import React from 'react';

import ConfigFilePreview from '@dash-frontend/components/ConfigFilePreview';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';

import styles from './PipelineSpec.module.css';

const PipelineSpec = () => {
  const {pipeline, loading} = useCurrentPipeline();

  return (
    <div className={styles.base}>
      {loading ? (
        <SkeletonBodyText lines={10} data-testid="PipelineSpec__loader" />
      ) : (
        <ConfigFilePreview config={JSON.parse(pipeline?.jsonSpec || '{}')} />
      )}
    </div>
  );
};

export default PipelineSpec;
