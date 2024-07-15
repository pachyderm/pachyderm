import React from 'react';

import ConfigFilePreview from '@dash-frontend/components/ConfigFilePreview';
import {useCurrentPipeline} from '@dash-frontend/hooks/useCurrentPipeline';
import {SkeletonBodyText} from '@pachyderm/components';

import styles from './PipelineSpec.module.css';

const PipelineSpec = () => {
  const {pipeline, loading} = useCurrentPipeline();

  return (
    <div className={styles.base}>
      {loading ? (
        <SkeletonBodyText lines={10} />
      ) : (
        <>
          <ConfigFilePreview
            allowMinimize
            title="Effective Spec"
            config={JSON.parse(pipeline?.effectiveSpecJson || '{}')}
            aria-label="Effective Spec"
            userSpecJSON={JSON.parse(pipeline?.userSpecJson || '{}')}
          />
          <ConfigFilePreview
            allowMinimize
            title="Submitted Spec"
            config={JSON.parse(pipeline?.userSpecJson || '{}')}
            aria-label="Submitted Spec"
          />
        </>
      )}
    </div>
  );
};

export default PipelineSpec;
