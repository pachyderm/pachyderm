import React from 'react';

import {RepoInfo} from '@dash-frontend/api/pfs';
import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import {
  TableViewPaginationWrapper,
  TableViewWrapper,
} from '@dash-frontend/components/TableView';
import {Table, LoadingDots, Pager} from '@pachyderm/components';

import {REPOS_DEFAULT_PAGE_SIZE} from '../../hooks/useRepoList';

import RepoListRow from './components/RepoListRow';
import styles from './RepositoriesList.module.css';

type RepositoriesListProps = {
  error?: string;
  loading: boolean;
  totalReposLength: number;
  repos?: RepoInfo[];
  pageIndex: number;
  updatePage: (page: number) => void;
  pageSize: number;
  setPageSize: React.Dispatch<React.SetStateAction<number>>;
  hasNextPage: boolean;
};

const RepositoriesList: React.FC<RepositoriesListProps> = ({
  totalReposLength,
  repos = [],
  loading,
  error,
  pageIndex,
  updatePage,
  pageSize,
  setPageSize,
  hasNextPage,
}) => {
  if (!loading && error) {
    return (
      <ErrorStateSupportLink
        title="We couldn't load the repositories list"
        message="Your repositories exist, but we couldn't fetch a list of them from our end. Please try refreshing this page."
      />
    );
  }

  if (!loading && totalReposLength === 0) {
    return (
      <EmptyState
        title="This DAG has no repositories"
        message="This is normal for new DAGs, but we still wanted to notify you because Pachyderm didn't detect any repositories on our end."
        linkToDocs={{
          text: 'Try creating a repository',
          pathWithoutDomain: 'concepts/data-concepts/repo/',
        }}
      />
    );
  }

  if (!loading && repos?.length === 0) {
    return (
      <EmptyState
        title="No matching results"
        message="We couldn't find any results matching your filters. Try using a different set of filters."
      />
    );
  }

  const boxShadowBottomOnly = pageIndex > 0 || hasNextPage;

  return (
    <>
      <TableViewWrapper>
        <Table>
          <Table.Head sticky>
            <Table.Row>
              <Table.HeaderCell boxShadowBottomOnly={boxShadowBottomOnly}>
                Name
              </Table.HeaderCell>
              <Table.HeaderCell boxShadowBottomOnly={boxShadowBottomOnly}>
                Size
              </Table.HeaderCell>
              <Table.HeaderCell boxShadowBottomOnly={boxShadowBottomOnly}>
                Created
              </Table.HeaderCell>
              <Table.HeaderCell boxShadowBottomOnly={boxShadowBottomOnly}>
                Last Commit
              </Table.HeaderCell>
              <Table.HeaderCell boxShadowBottomOnly={boxShadowBottomOnly}>
                Description
              </Table.HeaderCell>
              {repos?.[0]?.authInfo?.roles && (
                <Table.HeaderCell boxShadowBottomOnly={boxShadowBottomOnly}>
                  Roles
                </Table.HeaderCell>
              )}
              <Table.HeaderCell boxShadowBottomOnly={boxShadowBottomOnly} />
            </Table.Row>
          </Table.Head>

          <Table.Body>
            {loading && (
              <tr className={styles.loading}>
                <td>
                  <LoadingDots data-testid="RepositoriesList__loadingDots" />
                </td>
              </tr>
            )}
            {!loading &&
              repos.map((repo) => (
                <RepoListRow key={repo?.repo?.name} repo={repo} />
              ))}
          </Table.Body>
        </Table>
      </TableViewWrapper>
      {(pageIndex > 0 || hasNextPage) && (
        <TableViewPaginationWrapper>
          <Pager
            elementName="repo"
            page={pageIndex + 1}
            updatePage={updatePage}
            pageSizes={[REPOS_DEFAULT_PAGE_SIZE, 25, 50]}
            nextPageDisabled={!hasNextPage}
            updatePageSize={setPageSize}
            pageSize={pageSize}
          />
        </TableViewPaginationWrapper>
      )}
    </>
  );
};

export default RepositoriesList;
