import {SkeletonBodyText} from '@pachyderm/components';
import React from 'react';

import JsonSpec from '@dash-frontend/components/JsonSpec';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import styles from './PipelineSpec.module.css';

const PipelineSpec = () => {
  const {projectId} = useUrlState();
  const {pipeline, loading} = useCurrentPipeline();

  return (
    <div className={styles.base}>
      {loading ? (
        <SkeletonBodyText lines={10} data-testid="PipelineSpec__loader" />
      ) : (
        <JsonSpec
          jsonString={pipeline?.jsonSpec || '{}'}
          projectId={projectId}
        />
      )}
    </div>
  );
};

export default PipelineSpec;
