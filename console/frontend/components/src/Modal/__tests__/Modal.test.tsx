import {render, waitFor, act, screen} from '@testing-library/react';
import React from 'react';

import {click, tab} from '@dash-frontend/testHelpers';

import {Button} from './../../Button';
import {Modal, useModal} from './../../Modal';

const confirmMock = jest.fn();

const TestComponent = ({onShow}: {onShow?: () => void}) => {
  const {isOpen, closeModal, openModal} = useModal();

  return (
    <>
      <Modal show={isOpen} onHide={closeModal} onShow={onShow}>
        <Modal.Header>Header</Modal.Header>
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

    await waitFor(() =>
      expect(screen.queryByText('Header')).not.toBeInTheDocument(),
    );
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

    await waitFor(() =>
      expect(screen.queryByText('Header')).not.toBeInTheDocument(),
    );
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

    await waitFor(() =>
      expect(screen.queryByText('Header')).not.toBeInTheDocument(),
    );
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

  it('should keep focus within modal while tabbing', async () => {
    render(<TestComponent />);

    const openButton = screen.getByText('Open');
    await click(openButton);
    runTimers();

    await screen.findByText('Header');

    await tab();
    expect(screen.getByTestId('Modal__close')).toHaveFocus();

    await tab();
    expect(screen.getByRole('button', {name: /skip/i})).toHaveFocus();

    await tab();
    expect(screen.getByRole('button', {name: /confirm/i})).toHaveFocus();

    await tab();
    expect(screen.getByTestId('Modal__close')).toHaveFocus();
  });

  it('should call onShow on modal open', async () => {
    const onShow = jest.fn();
    render(<TestComponent onShow={onShow} />);

    const openButton = screen.getByText('Open');
    await click(openButton);
    runTimers();

    expect(onShow).toHaveBeenCalledTimes(1);
  });

  it('should close the modal on an escape key', async () => {
    render(<TestComponent />);

    const openButton = screen.getByText('Open');
    await click(openButton);
    runTimers();

    await screen.findByText('Header');

    act(() => {
      window.dispatchEvent(new KeyboardEvent('keydown', {key: 'Escape'}));
    });

    runTimers();

    await waitFor(() =>
      expect(screen.queryByText('Header')).not.toBeInTheDocument(),
    );
  });

  it('should close the modal on a  backdrop click', async () => {
    render(<TestComponent />);

    const openButton = screen.getByText('Open');
    await click(openButton);
    runTimers();

    await screen.findByText('Header');

    const backdrop = await screen.findByTestId('Modal__backdrop');
    await click(backdrop);

    runTimers();

    await waitFor(() =>
      expect(screen.queryByText('Header')).not.toBeInTheDocument(),
    );
  });
});
