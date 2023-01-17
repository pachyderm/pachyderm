import {render, waitFor, act, screen} from '@testing-library/react';
import React from 'react';

import {click} from '@dash-frontend/testHelpers';

import {Button} from './../../Button';
import {Modal, useModal} from './../../Modal';

const confirmMock = jest.fn();

const TestComponent = () => {
  const {isOpen, closeModal, openModal} = useModal();

  return (
    <>
      <Modal show={isOpen} onHide={closeModal}>
        <Modal.Header onHide={closeModal}>Header</Modal.Header>
        <Modal.Body>Body</Modal.Body>
        <Modal.Footer
          confirmText="Confirm"
          cancelText="Skip"
          onConfirm={confirmMock}
          onHide={closeModal}
        >
          Footer
        </Modal.Footer>
      </Modal>
      <Button onClick={openModal}>Open</Button>
      <Button onClick={closeModal}>Close</Button>
    </>
  );
};

const runTimers = () =>
  act(() => {
    jest.useFakeTimers();
    jest.runAllTimers();
    jest.useRealTimers();
  });

describe('Modal', () => {
  it('should be able to open and close the modal', async () => {
    render(<TestComponent />);

    const openButton = screen.getByText('Open');
    const closeButton = screen.getByText('Close');

    await click(openButton);
    runTimers();
    await screen.findByText('Header');

    await click(closeButton);

    runTimers();

    await waitFor(() => expect(screen.queryByText('Header')).toBeNull());
  });

  it('clicking the header close button should close the modal', async () => {
    render(<TestComponent />);

    const openButton = screen.getByText('Open');
    await click(openButton);
    runTimers();

    await screen.findByText('Header');

    const closeIcon = await screen.findByTestId('Modal__close');
    await click(closeIcon);

    runTimers();

    await waitFor(() => expect(screen.queryByText('Header')).toBeNull());
  });

  it('clicking the Cancel button should close the modal', async () => {
    render(<TestComponent />);

    const openButton = screen.getByText('Open');
    await click(openButton);
    runTimers();

    await screen.findByText('Header');

    const cancelButton = await screen.findByText('Skip');
    await click(cancelButton);

    runTimers();

    await waitFor(() => expect(screen.queryByText('Header')).toBeNull());
  });

  it('should call the confimation function when the confirm button is clicked', async () => {
    render(<TestComponent />);

    const openButton = screen.getByText('Open');
    await click(openButton);
    runTimers();

    const confirmButton = await screen.findByText('Confirm');

    await click(confirmButton);

    expect(confirmMock).toHaveBeenCalledTimes(1);
  });
});
