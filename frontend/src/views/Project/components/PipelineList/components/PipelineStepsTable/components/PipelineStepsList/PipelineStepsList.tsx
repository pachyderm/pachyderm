import React from 'react';

import {PipelineInfo} from '@dash-frontend/api/pps';
import EmptyState from '@dash-frontend/components/EmptyState';
import ErrorStateSupportLink from '@dash-frontend/components/ErrorStateSupportLink';
import {
  TableViewPaginationWrapper,
  TableViewWrapper,
} from '@dash-frontend/components/TableView';
import {Table, LoadingDots, Pager} from '@pachyderm/components';

import {PIPELINES_DEFAULT_PAGE_SIZE} from '../../hooks/usePipelineList';

import PipelineListRow from './components/PipelineListRow';
import styles from './PipelineStepsList.module.css';

type PipelineStepsListProps = {
  error?: string;
  loading: boolean;
  totalPipelinesLength: number;
  pipelines?: PipelineInfo[];
  pageIndex: number;
  updatePage: (page: number) => void;
  pageSize: number;
  setPageSize: React.Dispatch<React.SetStateAction<number>>;
  hasNextPage: boolean;
  isAuthActive: boolean;
};

const PipelineStepsList: React.FC<PipelineStepsListProps> = ({
  totalPipelinesLength,
  pipelines = [],
  loading,
  error,
  pageIndex,
  updatePage,
  pageSize,
  setPageSize,
  hasNextPage,
  isAuthActive,
}) => {
  if (!loading && error) {
    return (
      <ErrorStateSupportLink
        title="We couldn't load the pipelines list"
        message="Your pipelines exist, but we couldn't fetch a list of them from our end. Please try refreshing this page."
      />
    );
  }

  if (!loading && totalPipelinesLength === 0) {
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

  if (!loading && pipelines?.length === 0) {
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
                Pipeline State
              </Table.HeaderCell>
              <Table.HeaderCell boxShadowBottomOnly={boxShadowBottomOnly}>
                Last Job Status
              </Table.HeaderCell>
              <Table.HeaderCell boxShadowBottomOnly={boxShadowBottomOnly}>
                Version
              </Table.HeaderCell>
              <Table.HeaderCell boxShadowBottomOnly={boxShadowBottomOnly}>
                Created
              </Table.HeaderCell>
              <Table.HeaderCell boxShadowBottomOnly={boxShadowBottomOnly}>
                Description
              </Table.HeaderCell>
              {isAuthActive && (
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
                  <LoadingDots data-testid="PipelineStepsList__loadingDots" />
                </td>
              </tr>
            )}
            {!loading &&
              pipelines.map((pipeline) => {
                return (
                  pipeline && (
                    <PipelineListRow
                      key={pipeline.pipeline?.name}
                      pipeline={pipeline}
                    />
                  )
                );
              })}
          </Table.Body>
        </Table>
      </TableViewWrapper>
      {(pageIndex > 0 || hasNextPage) && (
        <TableViewPaginationWrapper>
          <Pager
            elementName="pipeline"
            page={pageIndex + 1}
            updatePage={updatePage}
            pageSizes={[PIPELINES_DEFAULT_PAGE_SIZE, 25, 50]}
            nextPageDisabled={!hasNextPage}
            updatePageSize={setPageSize}
            pageSize={pageSize}
          />
        </TableViewPaginationWrapper>
      )}
    </>
  );
};

export default PipelineStepsList;
