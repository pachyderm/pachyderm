import capitalize from 'lodash/capitalize';
import React, {useCallback, useEffect} from 'react';
import {useFieldArray, useForm} from 'react-hook-form';

import {ClusterPicker} from '@dash-frontend/api/metadata';
import {
  BranchPicker,
  CommitPicker,
  ProjectPicker,
  RepoPicker,
} from '@dash-frontend/api/pfs';
import BrandedDocLink from '@dash-frontend/components/BrandedDocLink';
import {useEditMetadata} from '@dash-frontend/hooks/useEditMetadata';
import useUrlState from '@dash-frontend/hooks/useUrlState';
import {
  Button,
  FormModal,
  Icon,
  Input,
  Table,
  TrashSVG,
} from '@pachyderm/components';

import {MetadataKeys} from '../../UserMetadata';

import styles from './EditMetadataModal.module.css';

type EditPicker =
  | ProjectPicker
  | CommitPicker
  | BranchPicker
  | RepoPicker
  | ClusterPicker;

type EditMetadataModalProps = {
  metadataType: MetadataKeys;
  metadata?: {
    key: string;
    value: string;
  }[];
  id?: string;
  closeModal: () => void;
  closeOnSuccess?: boolean;
};

const EditMetadataModal: React.FC<EditMetadataModalProps> = ({
  metadata,
  metadataType,
  id,
  closeModal,
  closeOnSuccess = true,
}) => {
  const {projectId, repoId} = useUrlState();
  const {updateMetadata, error} = useEditMetadata({
    onSuccess: closeOnSuccess ? closeModal : undefined,
  });

  const formCtx = useForm<{
    metadata: {id: string; key: string; value: string}[];
  }>();
  const {
    fields: envFields,
    append: appendEnv,
    remove: removeEnv,
  } = useFieldArray({
    control: formCtx.control,
    name: 'metadata',
  });

  const getMetadataResource = useCallback(
    (metadataType: MetadataKeys): EditPicker | null => {
      switch (metadataType) {
        case 'cluster':
          return {};
        case 'project':
          return {
            name: id,
          };
        case 'repo':
          return {
            name: {project: {name: projectId}, name: repoId, type: 'user'},
          };
        case 'commit':
          return {
            id: {
              id,
              repo: {
                name: {project: {name: projectId}, name: repoId, type: 'user'},
              },
            },
          };
        default:
          return null;
      }
    },
    [id, projectId, repoId],
  );

  useEffect(() => {
    // set metadata default fields
    formCtx.reset({
      metadata:
        metadata && metadata.length > 0 ? metadata : [{key: '', value: ''}],
    });
  }, [formCtx, metadata]);

  const handleSubmit = useCallback(
    ({metadata}: {metadata: {id: string; key: string; value: string}[]}) => {
      const metadataResource = getMetadataResource(metadataType);
      const newMetadata: {[key: string]: string} = {};

      metadata?.forEach(({key, value}) => {
        if (key || value) {
          newMetadata[key] = value;
        }
      });
      updateMetadata({
        [metadataType]: metadataResource,
        replace: {replacement: newMetadata},
      });
    },
    [getMetadataResource, metadataType, updateMetadata],
  );

  return (
    <FormModal
      isOpen={true}
      onHide={closeModal}
      headerText={`Edit ${capitalize(metadataType)} Metadata`}
      onSubmit={handleSubmit}
      formContext={formCtx}
      loading={false}
      error={error}
      confirmText="Apply Metadata"
      mode="Long"
      disabled={false}
      footerContent={
        <BrandedDocLink pathWithoutDomain="/set-up/metadata/">
          Info
        </BrandedDocLink>
      }
    >
      <Table>
        <Table.Head sticky>
          <Table.Row>
            <Table.HeaderCell className={styles.tableHeader}>
              key
            </Table.HeaderCell>
            <Table.HeaderCell className={styles.tableHeader}>
              value
            </Table.HeaderCell>
          </Table.Row>
        </Table.Head>

        <Table.Body>
          {(envFields as {id: string; key: string; value: string}[]).map(
            ({id, key, value}, index) => (
              <Table.Row key={id}>
                <Table.DataCell className={styles.tableField}>
                  <Input
                    className={styles.tableInput}
                    type="text"
                    placeholder="key"
                    defaultValue={key}
                    {...formCtx.register(`metadata.${index}.key`)}
                  />
                </Table.DataCell>
                <Table.DataCell className={styles.tableField}>
                  <Input
                    className={styles.tableInput}
                    type="text"
                    placeholder="value"
                    defaultValue={value}
                    {...formCtx.register(`metadata.${index}.value`)}
                  />
                  <Button
                    onClick={() => removeEnv(index)}
                    aria-label={`delete metadata row ${index}`}
                    type="button"
                    buttonType="ghost"
                  >
                    <Icon color="red" small>
                      <TrashSVG />
                    </Icon>
                  </Button>
                </Table.DataCell>
              </Table.Row>
            ),
          )}
          <tr>
            <td>
              <Button
                onClick={() => appendEnv({id: '', key: '', value: ''})}
                type="button"
                buttonType="secondary"
                className={styles.addNewButton}
              >
                Add new
              </Button>
            </td>
          </tr>
        </Table.Body>
      </Table>
    </FormModal>
  );
};

export default EditMetadataModal;
