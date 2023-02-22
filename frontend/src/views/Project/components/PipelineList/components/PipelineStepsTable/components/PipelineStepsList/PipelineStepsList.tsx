import {ApolloError} from '@apollo/client';
import {Pipeline} from '@graphqlTypes';
import {format, fromUnixTime} from 'date-fns';
import capitalize from 'lodash/capitalize';
import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import {TableViewWrapper} from '@dash-frontend/components/TableView';
import useUrlQueryState from '@dash-frontend/hooks/useUrlQueryState';
import {readableJobState} from '@dash-frontend/lib/jobs';
import readablePipelineState from '@dash-frontend/lib/readablePipelineState';
import {Table, LoadingDots} from '@pachyderm/components';

import usePipelineStepsList from './hooks/usePipelineStepsList';

type PipelineStepsListProps = {
  error?: ApolloError;
  loading: boolean;
  totalPipelinesLength: number;
  pipelines?: (Pipeline | null)[];
};

const getPipelineStateString = (nodeState: string, state: string) => {
  if (!nodeState || !state) {
    return '';
  } else if (state === nodeState) {
    return nodeState;
  } else {
    return `${nodeState} - ${state}`;
  }
};

const PipelineStepsList: React.FC<PipelineStepsListProps> = ({
  totalPipelinesLength,
  pipelines = [],
  loading,
  error,
}) => {
  const {iconItems, onOverflowMenuSelect} = usePipelineStepsList();
  const {viewState, toggleSelection} = useUrlQueryState();

  const addSelection = (value: string) => {
    toggleSelection('selectedPipelines', value);
  };

  if (loading) {
    return <LoadingDots />;
  }

  if (error) {
    return (
      <ErrorStateSupportLink
        title="We couldn't load the pipeline step list"
        message="Your pipeline steps exist, but we couldn't fetch a list of them from our end. Please try refreshing this page."
      />
    );
  }

  if (totalPipelinesLength === 0) {
    return (
      <EmptyState
        title="This DAG has no pipeline steps"
        message="This is normal for new DAGs, but we still wanted to notify you because Pachyderm didn't detect pipeline steps on our end."
        linkToDocs={{
          text: 'Try creating a pipeline',
          link: 'https://docs.pachyderm.com/latest/how-tos/pipeline-operations/create-pipeline/',
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
            <Table.HeaderCell>Pipeline Step</Table.HeaderCell>
            <Table.HeaderCell>Pipeline State</Table.HeaderCell>
            <Table.HeaderCell>Last Job Status</Table.HeaderCell>
            <Table.HeaderCell>Version</Table.HeaderCell>
            <Table.HeaderCell>Created</Table.HeaderCell>
            <Table.HeaderCell>Description</Table.HeaderCell>
            <Table.HeaderCell />
          </Table.Row>
        </Table.Head>

        <Table.Body>
          {pipelines.map((pipeline) => (
            <Table.Row
              key={pipeline?.id}
              data-testid="PipelineStepsList__row"
              onClick={() => addSelection(pipeline?.name || '')}
              isSelected={viewState.selectedPipelines?.includes(
                pipeline?.name || '',
              )}
              hasCheckbox
              overflowMenuItems={iconItems}
              dropdownOnSelect={onOverflowMenuSelect(pipeline?.name || '')}
            >
              <Table.DataCell>{pipeline?.name}</Table.DataCell>
              <Table.DataCell>
                {getPipelineStateString(
                  capitalize(pipeline?.nodeState || ''),
                  readablePipelineState(pipeline?.state || ''),
                )}
              </Table.DataCell>
              <Table.DataCell>
                {getPipelineStateString(
                  capitalize(pipeline?.lastJobNodeState || ''),
                  readableJobState(pipeline?.lastJobState || ''),
                )}
              </Table.DataCell>
              <Table.DataCell>v:{pipeline?.version}</Table.DataCell>
              <Table.DataCell>
                {pipeline?.createdAt
                  ? format(
                      fromUnixTime(pipeline?.createdAt),
                      'MMM dd, yyyy; h:mmaaa',
                    )
                  : '-'}
              </Table.DataCell>
              <Table.DataCell>{pipeline?.description || '-'}</Table.DataCell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table>
    </TableViewWrapper>
  );
};

export default PipelineStepsList;
