import React from 'react';

import Description from '@dash-frontend/components/Description';
import JSONBlock from '@dash-frontend/components/JSONBlock';
import PipelineInput from '@dash-frontend/components/PipelineInput';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import styles from './PipelineSpec.module.css';

const PipelineSpec = () => {
  const {projectId} = useUrlState();
  const {pipeline, loading} = useCurrentPipeline();

  return (
    <dl className={styles.base}>
      <Description term="Inputs" loading={loading} lines={9}>
        {pipeline && (
          <PipelineInput
            branchId="master"
            inputString={pipeline.inputString}
            projectId={projectId}
            className={styles.input}
          />
        )}
      </Description>

      {!loading && (
        <Description term="Transform">
          {pipeline?.transform ? (
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
      )}

      {!loading && (
        <Description term="Scheduling Spec">
          {pipeline?.schedulingSpec ? (
            <JSONBlock>
              {JSON.stringify(
                {
                  nodeSelectorMap: pipeline.schedulingSpec.nodeSelectorMap.map(
                    ({key, value}) => `${key}:${value}`,
                  ),
                  priorityClassName: pipeline.schedulingSpec.priorityClassName,
                },
                null,
                2,
              )}
            </JSONBlock>
          ) : (
            'N/A'
          )}
        </Description>
      )}
    </dl>
  );
};

export default PipelineSpec;
