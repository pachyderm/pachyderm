import React from 'react';

import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import {
  TableViewFilters,
  TableViewLoadingDots,
  TableViewPaginationWrapper,
  TableViewWrapper,
} from '@dash-frontend/components/TableView';
import {getStandardDateFromISOString} from '@dash-frontend/lib/dateTime';
import formatBytes from '@dash-frontend/lib/formatBytes';
import {Table, Form, Pager} from '@pachyderm/components';

import useCommitsList, {
  COMMITS_DEFAULT_PAGE_SIZE,
} from './hooks/useCommitsList';
import useCommitsListFilters from './hooks/useCommitsListFilters';

type CommitsListProps = {
  selectedRepo: string;
  filtersExpanded: boolean;
};

const CommitsList: React.FC<CommitsListProps> = ({
  selectedRepo,
  filtersExpanded,
}) => {
  const {formCtx, commitsFilters, staticFilterKeys, reverseOrder} =
    useCommitsListFilters();
  const {
    iconItems,
    onOverflowMenuSelect,
    loading,
    error,
    commits,
    cursors,
    updatePage,
    page,
    pageSize,
    setPageSize,
    hasNextPage,
  } = useCommitsList(selectedRepo, reverseOrder);

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
    <>
      <Form formContext={formCtx}>
        <TableViewFilters
          formCtx={formCtx}
          filtersExpanded={filtersExpanded}
          filters={commitsFilters}
          staticFilterKeys={staticFilterKeys}
        />
        {commits?.length === 0 ? (
          <EmptyState
            title="No matching results"
            message="We couldn't find any results matching your filters. Try using a different set of filters."
          />
        ) : (
          <TableViewWrapper hasPager>
            <Table>
              <Table.Head sticky>
                <Table.Row>
                  <Table.HeaderCell>Repository</Table.HeaderCell>
                  <Table.HeaderCell>Finished</Table.HeaderCell>
                  <Table.HeaderCell>ID</Table.HeaderCell>
                  <Table.HeaderCell>Branch</Table.HeaderCell>
                  <Table.HeaderCell>Size</Table.HeaderCell>
                  <Table.HeaderCell>Description</Table.HeaderCell>
                  <Table.HeaderCell />
                </Table.Row>
              </Table.Head>
              <Table.Body>
                {commits?.map((commit) => (
                  <Table.Row
                    key={`${commit.commit?.id}-${commit?.commit?.repo?.name}`}
                    data-testid="CommitsList__row"
                    overflowMenuItems={iconItems}
                    dropdownOnSelect={onOverflowMenuSelect(commit)}
                  >
                    <Table.DataCell>{`@${commit.commit?.repo?.name}`}</Table.DataCell>
                    <Table.DataCell width={210}>
                      {getStandardDateFromISOString(commit?.finished)}
                    </Table.DataCell>
                    <Table.DataCell width={330}>
                      {commit?.commit?.id}
                    </Table.DataCell>
                    <Table.DataCell>
                      {commit?.commit?.branch?.name}
                    </Table.DataCell>
                    <Table.DataCell width={120}>
                      {formatBytes(
                        commit?.details?.sizeBytes ??
                          commit?.sizeBytesUpperBound,
                      )}
                    </Table.DataCell>
                    <Table.DataCell>{commit.description}</Table.DataCell>
                  </Table.Row>
                ))}
              </Table.Body>
            </Table>
          </TableViewWrapper>
        )}
      </Form>
      {!loading &&
        commits &&
        commits?.length > 0 &&
        (hasNextPage || cursors.length > 1) && (
          <TableViewPaginationWrapper>
            <Pager
              elementName="file"
              page={page}
              updatePage={updatePage}
              pageSizes={[COMMITS_DEFAULT_PAGE_SIZE, 25, 50]}
              nextPageDisabled={!hasNextPage}
              updatePageSize={setPageSize}
              pageSize={pageSize}
            />
          </TableViewPaginationWrapper>
        )}
    </>
  );
};

export default CommitsList;
