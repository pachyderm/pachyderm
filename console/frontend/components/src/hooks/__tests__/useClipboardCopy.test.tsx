import {render, screen} from '@testing-library/react';
import React from 'react';

import {click} from '@dash-frontend/testHelpers';

import useClipboardCopy from '../useClipboardCopy';

Object.assign(navigator, {
  clipboard: {
    writeText: jest.fn(),
  },
});

const ClipboardComponent: React.FC = () => {
  const {copy, copied, reset, supported} = useClipboardCopy('copy text');

  if (!supported) return <div>Not Supported</div>;

  return (
    <>
      <div>{copied ? 'copied' : 'not copied'}</div>
      <div onClick={copy}>Copy</div>
      <div onClick={reset}>Reset</div>
    </>
  );
};

describe('useClipboardCopy', () => {
  it('should return a function that copies the text to the clipboard', async () => {
    render(<ClipboardComponent />);

    const copyButton = await screen.findByText('Copy');

    await click(copyButton);

    expect(navigator.clipboard.writeText).toHaveBeenCalledTimes(1);
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith('copy text');
  });

  it('should return a value indicating that the text has been copied', async () => {
    render(<ClipboardComponent />);

    const copyButton = await screen.findByText('Copy');
    const notCopied = await screen.findByText('not copied');
    expect(notCopied).toBeInTheDocument();

    await click(copyButton);

    const copied = await screen.findByText('copied');
    expect(copied).toBeInTheDocument();
  });

  it('should return a function to reset the copied state', async () => {
    render(<ClipboardComponent />);

    const copyButton = await screen.findByText('Copy');
    const resetButton = await screen.findByText('Reset');

    await click(copyButton);

    const copied = await screen.findByText('copied');
    expect(copied).toBeInTheDocument();

    await click(resetButton);
    const notCopied = await screen.findByText('not copied');
    expect(notCopied).toBeInTheDocument();
  });

  it('should return a value indicating if the copy command is supported', async () => {
    Object.assign(navigator, {
      clipboard: null,
    });

    render(<ClipboardComponent />);

    const notSupported = await screen.findByText('Not Supported');
    expect(notSupported).toBeInTheDocument();
  });
});
