import objectHash from 'object-hash';
import React, {useMemo} from 'react';

import {RepoInfo} from '@dash-frontend/api/pfs';
import Description from '@dash-frontend/components/Description';
import {PipelineLink} from '@dash-frontend/components/ResourceLink';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import {InputOutputNodesMap} from '@dash-frontend/lib/types';

import styles from './RepoInfo.module.css';

type RepoInfoProps = {
  repo?: RepoInfo;
  repoError?: string;
  currentRepoLoading: boolean;
  pipelineOutputsMap?: InputOutputNodesMap;
};

const RepoInfoSection: React.FC<RepoInfoProps> = ({
  repo,
  repoError,
  currentRepoLoading,
  pipelineOutputsMap = {},
}) => {
  const pipelineOutputs = useMemo(() => {
    const repoNodeName = objectHash({
      project: repo?.repo?.project?.name,
      name: repo?.repo?.name,
    });

    return pipelineOutputsMap[repoNodeName] || [];
  }, [pipelineOutputsMap, repo?.repo?.name, repo?.repo?.project?.name]);

  return (
    <>
      {repo?.description && (
        <div className={styles.description}>{repo?.description}</div>
      )}
      {pipelineOutputs.length > 0 && (
        <Description loading={currentRepoLoading} term="Inputs To">
          {pipelineOutputs.map(({name}) => (
            <PipelineLink name={name} key={name} />
          ))}
        </Description>
      )}
      <Description
        loading={currentRepoLoading}
        term="Repo Created"
        error={repoError}
      >
        {repo ? getStandardDateFromISOString(repo.created) : 'N/A'}
      </Description>
    </>
  );
};

export default RepoInfoSection;
