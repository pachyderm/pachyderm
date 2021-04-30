import {render} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';
import {mocked} from 'ts-jest/utils';

import useClipboardCopy from '../useClipboardCopy';

const queryCommandMock = mocked(window.document.queryCommandSupported);

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
    const {findByText} = render(<ClipboardComponent />);

    const copyButton = await findByText('Copy');

    userEvent.click(copyButton);

    expect(window.document.execCommand).toHaveBeenCalledTimes(1);
    expect(window.document.execCommand).toHaveBeenCalledWith('copy');
  });

  it('should return a value indicating that the text has been copied', async () => {
    const {findByText} = render(<ClipboardComponent />);

    const copyButton = await findByText('Copy');
    const notCopied = await findByText('not copied');
    expect(notCopied).toBeInTheDocument();

    userEvent.click(copyButton);

    const copied = await findByText('copied');
    expect(copied).toBeInTheDocument();
  });

  it('should return a function to reset the copied state', async () => {
    const {findByText} = render(<ClipboardComponent />);

    const copyButton = await findByText('Copy');
    const resetButton = await findByText('Reset');

    userEvent.click(copyButton);

    const copied = await findByText('copied');
    expect(copied).toBeInTheDocument();

    userEvent.click(resetButton);
    const notCopied = await findByText('not copied');
    expect(notCopied).toBeInTheDocument();
  });

  it('should return a value indicating if the copy command is supported', async () => {
    queryCommandMock.mockReturnValueOnce(false);

    const {findByText} = render(<ClipboardComponent />);

    const notSupported = await findByText('Not Supported');
    expect(notSupported).toBeInTheDocument();
  });
});
