import React from 'react';

import {RepoInfo} from '@dash-frontend/api/pfs';
import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import {TableViewWrapper} from '@dash-frontend/components/TableView';
import {Table, LoadingDots} from '@pachyderm/components';

import RepoListRow from './components/RepoListRow';

type RepositoriesListProps = {
  error?: string;
  loading: boolean;
  totalReposLength: number;
  repos?: RepoInfo[];
};

const RepositoriesList: React.FC<RepositoriesListProps> = ({
  totalReposLength,
  repos = [],
  loading,
  error,
}) => {
  if (loading) {
    return <LoadingDots />;
  }

  if (error) {
    return (
      <ErrorStateSupportLink
        title="We couldn't load the repositories list"
        message="Your repositories exist, but we couldn't fetch a list of them from our end. Please try refreshing this page."
      />
    );
  }

  if (totalReposLength === 0) {
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

  if (repos?.length === 0) {
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
            <Table.HeaderCell>Size</Table.HeaderCell>
            <Table.HeaderCell>Created</Table.HeaderCell>
            <Table.HeaderCell>Last Commit</Table.HeaderCell>
            <Table.HeaderCell>Description</Table.HeaderCell>
            {repos?.[0]?.authInfo?.roles && (
              <Table.HeaderCell>Roles</Table.HeaderCell>
            )}
            <Table.HeaderCell />
          </Table.Row>
        </Table.Head>

        <Table.Body>
          {repos.map((repo) => (
            <RepoListRow key={repo?.repo?.name} repo={repo} />
          ))}
        </Table.Body>
      </Table>
    </TableViewWrapper>
  );
};

export default RepositoriesList;
