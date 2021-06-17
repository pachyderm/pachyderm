import {render, waitFor, act} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

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

describe('Modal', () => {
  it('should be able to open and close the modal', async () => {
    const {findByText, getByText, queryByText} = render(<TestComponent />);

    const openButton = getByText('Open');
    const closeButton = getByText('Close');

    userEvent.click(openButton);
    await findByText('Header');

    userEvent.click(closeButton);

    act(() => {
      jest.runAllTimers();
    });

    await waitFor(() => expect(queryByText('Header')).toBeNull());
  });

  it('clicking the header close button should close the modal', async () => {
    const {findByText, getByText, findByTestId, queryByText} = render(
      <TestComponent />,
    );

    const openButton = getByText('Open');
    userEvent.click(openButton);

    await findByText('Header');

    const closeIcon = await findByTestId('Modal__close');
    userEvent.click(closeIcon);

    act(() => {
      jest.runAllTimers();
    });

    await waitFor(() => expect(queryByText('Header')).toBeNull());
  });

  it('clicking the Cancel button should close the modal', async () => {
    const {findByText, getByText, queryByText} = render(<TestComponent />);

    const openButton = getByText('Open');
    userEvent.click(openButton);

    await findByText('Header');

    const cancelButton = await findByText('Cancel');
    userEvent.click(cancelButton);

    act(() => {
      jest.runAllTimers();
    });

    await waitFor(() => expect(queryByText('Header')).toBeNull());
  });

  it('should call the confimation function when the confirm button is clicked', async () => {
    const {findByText, getByText} = render(<TestComponent />);

    const openButton = getByText('Open');
    userEvent.click(openButton);

    const confirmButton = await findByText('Confirm');

    userEvent.click(confirmButton);

    expect(confirmMock).toHaveBeenCalledTimes(1);
  });
});
