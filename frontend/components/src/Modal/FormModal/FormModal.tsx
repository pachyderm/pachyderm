import noop from 'lodash/noop';
import React from 'react';
import {UseFormReturn, SubmitHandler, FieldValues} from 'react-hook-form';

import {LoadingDots} from '@pachyderm/components';

import Modal from '../components/Modal';
import {ModalModes} from '../components/Modal/Modal';
import ModalBody from '../components/ModalBody';
import ModalFooter from '../components/ModalFooter';
import ModalHeader from '../components/ModalHeader';
import ModalStatus from '../components/ModalStatus';

import {Form} from './../../Form';
import styles from './FormModal.module.css';

export interface FormModalProps<T extends FieldValues> {
  children?: React.ReactNode;
  isOpen?: boolean;
  onHide?: () => void;
  error?: string;
  loading?: boolean;
  updating?: boolean;
  formContext: UseFormReturn<T>;
  onSubmit: SubmitHandler<T>;
  confirmText?: string;
  headerText?: string;
  success?: boolean;
  mode?: ModalModes;
  disabled?: boolean;
  footerContent?: JSX.Element;
}

const FormModal = <T extends FieldValues>({
  children,
  isOpen = false,
  onHide = noop,
  error = '',
  loading,
  updating,
  formContext,
  onSubmit,
  confirmText = 'Submit',
  headerText,
  success,
  mode = 'Default',
  disabled,
  footerContent,
}: FormModalProps<T>) => {
  const modalStatus =
    (updating && 'updating') ||
    (error && 'error') ||
    (success && 'success') ||
    null;

  const modalStatusMessage = error || (success && "You're all set!") || '';

  return (
    <Modal show={isOpen} onHide={onHide} mode={mode}>
      <Form
        className={styles.form}
        formContext={formContext}
        onSubmit={() => {
          formContext.handleSubmit(onSubmit)();
        }}
      >
        {modalStatus ? (
          <ModalStatus status={modalStatus}>{modalStatusMessage}</ModalStatus>
        ) : null}

        <ModalHeader>{headerText}</ModalHeader>

        <ModalBody>{loading ? <LoadingDots /> : children}</ModalBody>

        <ModalFooter
          buttonType="submit"
          confirmText={confirmText}
          disabled={disabled}
          onHide={onHide}
          footerContent={footerContent}
        />
      </Form>
    </Modal>
  );
};

export default FormModal;
