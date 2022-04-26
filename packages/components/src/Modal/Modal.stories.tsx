import React, {useState} from 'react';

import {Button} from './../Button';

import {BasicModal, useModal, WizardModal, FullPageModal} from './';

export default {
  title: 'Modal',
  component: BasicModal,
  argTypes: {
    loading: {
      defaultValue: false,
    },
  },
};

export const Default = () => {
  const {isOpen, openModal, closeModal} = useModal();

  return (
    <>
      <Button onClick={openModal}>Open Modal</Button>
      <BasicModal
        show={isOpen}
        onHide={closeModal}
        headerContent="Non Actionable Modal"
        loading={false}
        small={true}
      >
        The default Modal is meant for Modals that provide information or
        instructions but to not require actions from the user. Users are able to
        dismiss the modal by clicking the background of the screen.
      </BasicModal>
    </>
  );
};

export const LongContentWithControls: React.FC = (modalArgs) => {
  const {isOpen, openModal, closeModal} = useModal();

  return (
    <>
      <Button onClick={openModal}>Open Modal</Button>
      <BasicModal
        {...modalArgs}
        show={isOpen}
        onHide={closeModal}
        headerContent="Long text that should overflow and wrap to the next line"
      >
        Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod
        tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim
        veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea
        commodo consequat. Duis aute irure dolor in reprehenderit in voluptate
        velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint
        occaecat cupidatat non proident, sunt in culpa qui officia deserunt
        mollit anim id est laborum. Sed ut perspiciatis unde omnis iste natus
        error sit voluptatem accusantium doloremque laudantium, totam rem
        aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto
        beatae vitae dicta sunt explicabo. Nemo enim ipsam voluptatem quia
        voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni
        dolores eos qui ratione voluptatem sequi nesciunt. Neque porro quisquam
        est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit,
        sed quia non numquam eius modi tempora incidunt ut labore et dolore
        magnam aliquam quaerat voluptatem. Ut enim ad minima veniam, quis
        nostrum exercitationem ullam corporis suscipit laboriosam, nisi ut
        aliquid ex ea commodi consequatur? Quis autem vel eum iure reprehenderit
        qui in ea voluptate velit esse quam nihil molestiae consequatur, vel
        illum qui dolorem eum fugiat quo voluptas nulla pariatur?
      </BasicModal>
    </>
  );
};

export const Actionable = () => {
  const {isOpen, openModal, closeModal} = useModal();
  const [success, setSuccess] = useState(false);

  return (
    <>
      <Button onClick={openModal}>Open Modal</Button>
      <BasicModal
        successMessage={success ? "You're all set!" : undefined}
        show={isOpen}
        onHide={closeModal}
        headerContent="Actionable Modal"
        actionable
        onConfirm={() => setSuccess(true)}
        confirmText="Confirm"
        loading={false}
        small={true}
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
      <Button onClick={openModal}>Open Modal</Button>
      <BasicModal
        show={isOpen}
        onHide={closeModal}
        headerContent="Error Modal"
        hideActions
        loading={false}
        actionable
        errorMessage="Promo has expired on 01/01/2020."
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
      <Button onClick={openModal}>Open Modal</Button>
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
      <Button onClick={openModal}>Open Modal</Button>
      <FullPageModal show={isOpen} onHide={closeModal} hideType="exit">
        This is a full page modal
      </FullPageModal>
    </>
  );
};
