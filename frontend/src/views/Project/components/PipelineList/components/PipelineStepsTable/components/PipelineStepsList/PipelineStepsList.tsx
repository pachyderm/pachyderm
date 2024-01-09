import React from 'react';

import {RepoInfo} from '@dash-frontend/api/pfs';
import {PipelineInfo} from '@dash-frontend/api/pps';
import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import {TableViewWrapper} from '@dash-frontend/components/TableView';
import {Table, LoadingDots} from '@pachyderm/components';

import PipelineListRow from './components/PipelineListRow';

type PipelineStepsListProps = {
  error?: string;
  loading: boolean;
  totalPipelinesLength: number;
  pipelines?: PipelineInfo[];
  pipelineRepoMap: Record<string, RepoInfo>;
};

const PipelineStepsList: React.FC<PipelineStepsListProps> = ({
  totalPipelinesLength,
  pipelines = [],
  loading,
  error,
  pipelineRepoMap,
}) => {
  if (loading) {
    return <LoadingDots />;
  }

  if (error) {
    return (
      <ErrorStateSupportLink
        title="We couldn't load the pipelines list"
        message="Your pipelines exist, but we couldn't fetch a list of them from our end. Please try refreshing this page."
      />
    );
  }

  if (totalPipelinesLength === 0) {
    return (
      <EmptyState
        title="This DAG has no pipelines"
        message="This is normal for new DAGs, but we still wanted to notify you because Pachyderm didn't detect any pipelines on our end."
        linkToDocs={{
          text: 'Try creating a pipeline',
          pathWithoutDomain: 'how-tos/pipeline-operations/create-pipeline/',
        }}
      />
    );
  }

  if (pipelines?.length === 0) {
    return (
      <EmptyState
        title="No matching results"
        message="We couldn't find any results matching your filters. Try using a different set of filters."
      />
    );
  }

  return (
    <TableViewWrapper>
      <Table>
        <Table.Head sticky>
          <Table.Row>
            <Table.HeaderCell>Name</Table.HeaderCell>
            <Table.HeaderCell>Pipeline State</Table.HeaderCell>
            <Table.HeaderCell>Last Job Status</Table.HeaderCell>
            <Table.HeaderCell>Version</Table.HeaderCell>
            <Table.HeaderCell>Created</Table.HeaderCell>
            <Table.HeaderCell>Description</Table.HeaderCell>
            {pipelineRepoMap[pipelines?.[0]?.pipeline?.name || '']?.authInfo
              ?.roles && <Table.HeaderCell>Roles</Table.HeaderCell>}
            <Table.HeaderCell />
          </Table.Row>
        </Table.Head>
        <Table.Body>
          {pipelines.map((pipeline) => {
            return (
              pipeline && (
                <PipelineListRow
                  key={pipeline.pipeline?.name}
                  pipeline={pipeline}
                  pipelineRepoMap={pipelineRepoMap}
                />
              )
            );
          })}
        </Table.Body>
      </Table>
    </TableViewWrapper>
  );
};

export default PipelineStepsList;
