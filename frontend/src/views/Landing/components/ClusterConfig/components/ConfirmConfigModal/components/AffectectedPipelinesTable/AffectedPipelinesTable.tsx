import {SetClusterDefaultsResp, PipelineObject} from '@graphqlTypes';
import React from 'react';

import {
  lineageRoute,
  pipelineRoute,
} from '@dash-frontend/views/Project/utils/routes';
import {Table, useSort, stringComparator, Link} from '@pachyderm/components';

import styles from './AffectedPipelinesTable.module.css';

type AffectedPipelinesTableProps = {
  affectedPipelines: SetClusterDefaultsResp['affectedPipelinesList'];
};

const projectComparator = {
  name: 'Project',
  func: stringComparator,
  accessor: (pipeline: PipelineObject | null) => pipeline?.project?.name || '',
};

const nameComparator = {
  name: 'Name',
  func: stringComparator,
  accessor: (pipeline: PipelineObject | null) => pipeline?.name || '',
};

export const AffectedPipelinesTable: React.FC<AffectedPipelinesTableProps> = ({
  affectedPipelines,
}) => {
  const {sortedData, setComparator, reversed, comparatorName} = useSort({
    data: affectedPipelines,
    initialSort: projectComparator,
    initialDirection: 1,
  });

  return (
    <div className={styles.base}>
      <Table>
        <Table.Head sticky>
          <Table.Row>
            <Table.HeaderCell
              onClick={() => setComparator(projectComparator)}
              sortable={true}
              sortLabel="project"
              sortSelected={comparatorName === 'Project'}
              sortReversed={!reversed}
              className={styles.tableCell}
            >
              Project
            </Table.HeaderCell>
            <Table.HeaderCell
              onClick={() => {
                setComparator(nameComparator);
              }}
              sortable={true}
              sortLabel="name"
              sortSelected={comparatorName === 'Name'}
              sortReversed={!reversed}
              className={styles.tableCell}
            >
              Name
            </Table.HeaderCell>
          </Table.Row>
        </Table.Head>

        <Table.Body>
          {sortedData.map((pipeline, i) => {
            const projectNameAlreadyListed =
              i > 0 &&
              pipeline?.project?.name === sortedData[i - 1]?.project?.name;
            return (
              <Table.Row key={`${pipeline?.project?.name}_${pipeline?.name}`}>
                <Table.DataCell className={styles.tableCell}>
                  <Link
                    target="_blank"
                    rel="noopener noreferrer"
                    to={lineageRoute({
                      projectId: pipeline?.project?.name || '',
                    })}
                  >
                    {projectNameAlreadyListed ? '' : pipeline?.project?.name}
                  </Link>
                </Table.DataCell>
                <Table.DataCell className={styles.tableCell}>
                  <Link
                    target="_blank"
                    rel="noopener noreferrer"
                    to={pipelineRoute({
                      projectId: pipeline?.project?.name || '',
                      pipelineId: pipeline?.name || '',
                    })}
                  >
                    {pipeline?.name}
                  </Link>
                </Table.DataCell>
              </Table.Row>
            );
          })}
        </Table.Body>
      </Table>
    </div>
  );
};

export default AffectedPipelinesTable;
