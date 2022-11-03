import noop from 'lodash/noop';
import React from 'react';
import {UseFormReturn, SubmitHandler, FieldValues} from 'react-hook-form';

import {LoadingDots} from 'LoadingDots';

import Modal from '../components/Modal';
import ModalBody from '../components/ModalBody';
import ModalFooter from '../components/ModalFooter';
import ModalHeader from '../components/ModalHeader';
import ModalStatus from '../components/ModalStatus';

import {Form} from './../../Form';

export interface FormModalProps<T extends FieldValues> {
  children: React.ReactNode;
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
  small?: boolean;
  disabled?: boolean;
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
  small,
  disabled,
}: FormModalProps<T>) => {
  const modalStatus =
    (updating && 'updating') ||
    (error && 'error') ||
    (success && 'success') ||
    null;

  const modalStatusMessage = error || (success && "You're all set!") || '';

  return (
    <Form formContext={formContext} onSubmit={onSubmit}>
      <Modal show={isOpen} onHide={onHide} small={small} actionable>
        {modalStatus ? (
          <ModalStatus status={modalStatus}>{modalStatusMessage}</ModalStatus>
        ) : null}

        <ModalHeader onHide={onHide} actionable>
          {headerText}
        </ModalHeader>

        <ModalBody>{loading ? <LoadingDots /> : children}</ModalBody>

        <ModalFooter
          buttonType="submit"
          confirmText={confirmText}
          disabled={disabled}
          onConfirm={() => {
            formContext.handleSubmit(onSubmit)();
          }}
          onHide={onHide}
        />
      </Modal>
    </Form>
  );
};

export default FormModal;
