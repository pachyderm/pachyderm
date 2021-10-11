import React, {useState} from 'react';

import {Button} from './../Button';

import {BasicModal, useModal, WizardModal, FullPageModal} from './';

export default {title: 'Modal'};

export const Default = () => {
  const {isOpen, openModal, closeModal} = useModal();

  return (
    <>
      <Button autoWidth autoHeight onClick={openModal}>
        Open Modal
      </Button>
      <BasicModal
        show={isOpen}
        onHide={closeModal}
        headerContent="Non Actionable Modal"
        loading={false}
      >
        The default Modal is meant for Modals that provide information or
        instructions but to not require actions from the user. Users are able to
        dismiss the modal by clicking the background of the screen.
      </BasicModal>
    </>
  );
};

export const Actionable = () => {
  const {isOpen, openModal, closeModal} = useModal();
  const [success, setSuccess] = useState(false);

  return (
    <>
      <Button autoWidth autoHeight onClick={openModal}>
        Open Modal
      </Button>
      <BasicModal
        success={success}
        show={isOpen}
        onHide={closeModal}
        headerContent="Actionable Modal"
        actionable
        onConfirm={() => setSuccess(true)}
        confirmText="Confirm"
        loading={false}
      >
        The actionable modal is for modals that require input or actions from
        the user. Users are not able to close the Modal by clicking the screen
        background. When the action is completed by clicking onConfirm, success
        will be shown by the appearance of a green check mark in the header.
      </BasicModal>
    </>
  );
};

export const ErrorState = () => {
  const {isOpen, openModal, closeModal} = useModal();

  return (
    <>
      <Button autoWidth autoHeight onClick={openModal}>
        Open Modal
      </Button>
      <BasicModal
        show={isOpen}
        onHide={closeModal}
        headerContent="Error Modal"
        hideActions
        loading={false}
        actionable
        errorMessage="Promo has expired on 01/01/2020. Please contact us for new promo."
      >
        This Modal has an error
      </BasicModal>
    </>
  );
};

export const Wizard = () => {
  const {isOpen, openModal, closeModal} = useModal();

  return (
    <>
      <Button autoWidth autoHeight onClick={openModal}>
        Open Modal
      </Button>
      <WizardModal
        show={isOpen}
        onHide={closeModal}
        headerContent={['Wizard Modal Page 1', 'Wizard Modal Page 2']}
        loading={false}
        modalContent={[
          'The wizard modal is for modals that have multiple pages, and navigation between those pages. This is content for page 1.',
          "This is page 2 of the modal. Pressing 'done' simply closes the modal",
        ]}
      />
    </>
  );
};

export const FullPage = () => {
  const {isOpen, openModal, closeModal} = useModal();

  return (
    <>
      <Button autoWidth autoHeight onClick={openModal}>
        Open Modal
      </Button>
      <FullPageModal show={isOpen} onHide={closeModal}>
        This is a full page modal
      </FullPageModal>
    </>
  );
};
