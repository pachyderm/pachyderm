import React from 'react';

import JSONBlock from '@dash-frontend/components/JSONBlock';
import PipelineInput from '@dash-frontend/components/PipelineInput';
import useCurrentPipeline from '@dash-frontend/hooks/useCurrentPipeline';
import useUrlState from '@dash-frontend/hooks/useUrlState';

import Description from '../../../Description';

import styles from './PipelineSpec.module.css';

const PipelineSpec = () => {
  const {projectId} = useUrlState();
  const {pipeline, loading} = useCurrentPipeline();

  return (
    <dl>
      <Description term="Inputs" loading={loading} lines={5}>
        {pipeline && (
          <PipelineInput
            inputString={pipeline.inputString}
            projectId={projectId}
            className={styles.input}
          />
        )}
      </Description>

      <Description term="Transform" loading={loading} lines={3}>
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

      <Description term="Scheduling Spec" loading={loading} lines={5}>
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
    </dl>
  );
};

export default PipelineSpec;
