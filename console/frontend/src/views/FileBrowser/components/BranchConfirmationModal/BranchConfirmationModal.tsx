import {BranchInfo} from '@dash-frontend/generated/proto/pfs/pfs.pb';
import React from 'react';
import {useForm} from 'react-hook-form';

import {FormModal, Select} from '@pachyderm/components';

import styles from './BranchConfirmationModal.module.css';

type BranchConfirmationModalProps = {
  commitBranches?: BranchInfo[];
  onHide: () => void;
  onSubmit: (formData: BranchFormFields) => void;
  deleteError?: string;
  loading?: boolean;
  children?: React.ReactNode;
};

type BranchFormFields = {
  branch: string;
};

const BranchConfirmationModal = ({
  onHide,
  commitBranches,
  onSubmit,
  deleteError,
  loading,
  children,
}: BranchConfirmationModalProps) => {
  const branchSelectionFormCtx = useForm<BranchFormFields>({
    mode: 'onChange',
  });

  return (
    <FormModal
      isOpen={true}
      onHide={onHide}
      confirmText="Confirm"
      formContext={branchSelectionFormCtx}
      headerText="Multiple branches found"
      onSubmit={onSubmit}
      loading={loading}
      error={
        deleteError || branchSelectionFormCtx.formState.errors.branch?.message
      }
    >
      <div className={styles.branchSelection}>
        This commit is linked to multiple branches. Please select the branch
        where you want to perform the deletion:
      </div>
      <Select
        id="branch"
        placeholder="Select a branch"
        validationOptions={{
          required: {
            value: true,
            message: 'Please select a branch',
          },
        }}
      >
        {commitBranches?.map((branch) => {
          const branchName = branch.branch?.name || '';
          return (
            <Select.Option key={branchName} value={branchName}>
              {branchName}
            </Select.Option>
          );
        })}
      </Select>
      {children && (
        <div className={styles.textContent}>Deleting {children}</div>
      )}
    </FormModal>
  );
};

export default BranchConfirmationModal;
