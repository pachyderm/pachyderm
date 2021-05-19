import noop from 'lodash/noop';
import React from 'react';
import {UseFormReturn, SubmitHandler} from 'react-hook-form';

import Modal from '../components/Modal';
import ModalBody from '../components/ModalBody';
import ModalError from '../components/ModalError';
import ModalFooter from '../components/ModalFooter';
import ModalHeader from '../components/ModalHeader';

import {Form} from './../../Form';

export interface FormModalProps<T> {
  children: React.ReactNode;
  isOpen?: boolean;
  onHide?: () => void;
  error?: string;
  loading?: boolean;
  success?: boolean;
  formContext: UseFormReturn<T>;
  onSubmit: SubmitHandler<T>;
  confirmText?: string;
  headerText?: string;
}

const FormModal = <T,>({
  children,
  isOpen = false,
  onHide = noop,
  error = '',
  loading,
  success,
  formContext,
  onSubmit,
  confirmText = 'Submit',
  headerText,
}: FormModalProps<T>) => {
  return (
    <Modal show={isOpen} onHide={onHide} actionable>
      {error && <ModalError>{error}</ModalError>}

      <ModalHeader
        onHide={onHide}
        loading={loading}
        success={success}
        actionable
      >
        {headerText}
      </ModalHeader>

      <Form formContext={formContext} onSubmit={onSubmit}>
        <ModalBody>{children}</ModalBody>

        <ModalFooter
          buttonType="submit"
          confirmText={confirmText}
          disabled={loading}
          onConfirm={noop}
          onHide={onHide}
        />
      </Form>
    </Modal>
  );
};

export default FormModal;
