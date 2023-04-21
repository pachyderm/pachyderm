import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import {
  TableViewFilters,
  TableViewLoadingDots,
} from '@dash-frontend/components/TableView';
import useCommits, {COMMIT_LIMIT} from '@dash-frontend/hooks/useCommits';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {getStandardDate} from '@dash-frontend/lib/dateTime';
import {Table, Form} from '@pachyderm/components';

import useCommitsList from './hooks/useCommitsList';
import useCommitsListFilters, {
  commitsFilters,
} from './hooks/useCommitsListFilters';

type CommitsListProps = {
  selectedRepo: string;
  filtersExpanded: boolean;
};

const CommitsList: React.FC<CommitsListProps> = ({
  selectedRepo,
  filtersExpanded,
}) => {
  const {projectId} = useUrlState();
  const {commits, loading, error} = useCommits({
    args: {
      projectId,
      repoName: selectedRepo,
      number: COMMIT_LIMIT,
    },
  });
  const {
    sortedCommits,
    formCtx,
    staticFilterKeys,
    clearableFiltersMap,
    multiselectFilters,
  } = useCommitsListFilters({commits});
  const {iconItems, onOverflowMenuSelect} = useCommitsList();

  if (loading) {
    return <TableViewLoadingDots data-testid="CommitsList__loadingDots" />;
  }

  if (error) {
    return (
      <ErrorStateSupportLink
        title="We couldn't load the commits list"
        message="Your commits have been processed, but we couldn't fetch a list of them from our end. Please try refreshing this page."
      />
    );
  }

  return (
    <Form formContext={formCtx}>
      <TableViewFilters
        formCtx={formCtx}
        filtersExpanded={filtersExpanded}
        filters={commitsFilters}
        multiselectFilters={multiselectFilters}
        clearableFiltersMap={clearableFiltersMap}
        staticFilterKeys={staticFilterKeys}
      />
      {sortedCommits?.length === 0 ? (
        <EmptyState
          title="No matching results"
          message="We couldn't find any results matching your filters. Try using a different set of filters."
        />
      ) : (
        <Table>
          <Table.Head sticky>
            <Table.Row>
              <Table.HeaderCell>Repository</Table.HeaderCell>
              <Table.HeaderCell>Finished</Table.HeaderCell>
              <Table.HeaderCell>ID</Table.HeaderCell>
              <Table.HeaderCell>Branch</Table.HeaderCell>
              <Table.HeaderCell>Size</Table.HeaderCell>
              <Table.HeaderCell>Origin</Table.HeaderCell>
              <Table.HeaderCell>Description</Table.HeaderCell>
              <Table.HeaderCell />
            </Table.Row>
          </Table.Head>

          <Table.Body>
            {sortedCommits.map((commit) => (
              <Table.Row
                key={`${commit.id}-${commit.repoName}`}
                data-testid="CommitsList__row"
                overflowMenuItems={iconItems}
                dropdownOnSelect={onOverflowMenuSelect(commit)}
              >
                <Table.DataCell>{`@${commit.repoName}`}</Table.DataCell>
                <Table.DataCell>
                  {getStandardDate(commit?.finished)}
                </Table.DataCell>
                <Table.DataCell>{commit?.id.slice(0, 6)}...</Table.DataCell>
                <Table.DataCell>{commit?.branch?.name || '-'}</Table.DataCell>
                <Table.DataCell>{commit.sizeDisplay}</Table.DataCell>
                <Table.DataCell>{commit.originKind}</Table.DataCell>
                <Table.DataCell>{commit.description}</Table.DataCell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table>
      )}
    </Form>
  );
};

export default CommitsList;
