import capitalize from 'lodash/capitalize';
import isEmpty from 'lodash/isEmpty';
import React, {useMemo} from 'react';

import CollapsibleSection from '@dash-frontend/components/CollapsibleSection';
import {
  Button,
  CaptionTextSmall,
  EditSVG,
  Icon,
  InfoSVG,
  Table,
  stringComparator,
  useModal,
  useSort,
} from '@pachyderm/components';

import EditMetadataModal from './components/EditMetadataModal';
import styles from './UserMetadata.module.css';

export type MetadataKeys =
  | 'repo'
  | 'pipeline'
  | 'commit'
  | 'job'
  | 'project'
  | 'cluster'
  | 'branch';

type UserMetadataProps = {
  metadataType: MetadataKeys;
  metadata?: {[key: string]: string};
  id?: string;
  captionText?: string;
  editable?: boolean;
};

const keyComparator = {
  name: 'Key',
  func: stringComparator,
  accessor: (metadata: {key: string; value: string}) => metadata.key,
};

const valueComparator = {
  name: 'Value',
  func: stringComparator,
  accessor: (metadata: {key: string; value: string}) => metadata.value,
};

const UserMetadata: React.FC<UserMetadataProps> = ({
  metadata,
  metadataType,
  id,
  captionText,
  editable = false,
}) => {
  const {isOpen, openModal, closeModal} = useModal();

  const metadataArray = useMemo(() => {
    return !isEmpty(metadata)
      ? Object.keys(metadata).map((key) => ({
          key,
          value: metadata[key],
        }))
      : [];
  }, [metadata]);

  const {
    sortedData: sortedMetadata,
    setComparator,
    reversed,
    comparatorName,
  } = useSort({
    data: metadataArray,
    initialSort: keyComparator,
    initialDirection: 1,
  });

  return (
    <div className={styles.base}>
      <CollapsibleSection
        boxShadow
        header={
          <div className={styles.topRow}>
            {capitalize(metadataType)} Metadata
            {editable && (
              <Button
                buttonType="ghost"
                onClick={(e) => {
                  e.stopPropagation();
                  openModal();
                }}
                IconSVG={EditSVG}
                iconPosition="end"
              >
                Edit
              </Button>
            )}
          </div>
        }
        maxHeight={300}
      >
        {captionText && (
          <CaptionTextSmall className={styles.captionText}>
            {captionText}
          </CaptionTextSmall>
        )}
        <Table>
          <Table.Head sticky>
            <Table.Row>
              <Table.HeaderCell
                onClick={() => setComparator(keyComparator)}
                sortable={sortedMetadata.length > 0}
                sortLabel="key"
                sortSelected={comparatorName === 'Key'}
                sortReversed={!reversed}
              >
                key
              </Table.HeaderCell>
              <Table.HeaderCell
                onClick={() => setComparator(valueComparator)}
                sortable={sortedMetadata.length > 0}
                sortLabel="value"
                sortSelected={comparatorName === 'Value'}
                sortReversed={!reversed}
              >
                value
              </Table.HeaderCell>
            </Table.Row>
          </Table.Head>

          <Table.Body>
            {sortedMetadata.length > 0 ? (
              sortedMetadata.map(({key, value}) => {
                return (
                  <Table.Row key={key}>
                    <Table.DataCell className={styles.tableCell}>
                      {key}
                    </Table.DataCell>
                    <Table.DataCell className={styles.tableCell}>
                      {value}
                    </Table.DataCell>
                  </Table.Row>
                );
              })
            ) : (
              <Table.Row>
                <Table.DataCell className={styles.tableCell}>
                  <CaptionTextSmall>
                    <Icon small>
                      <InfoSVG />
                    </Icon>{' '}
                    No user defined metadata
                  </CaptionTextSmall>
                </Table.DataCell>
              </Table.Row>
            )}
          </Table.Body>
        </Table>
      </CollapsibleSection>
      {isOpen && (
        <EditMetadataModal
          metadata={sortedMetadata}
          metadataType={metadataType}
          id={id}
          closeModal={closeModal}
        />
      )}
    </div>
  );
};

export default UserMetadata;
