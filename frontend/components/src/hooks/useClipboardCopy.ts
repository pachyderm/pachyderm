import {useState, useEffect, useCallback} from 'react';

const useClipboardCopy = (copyText: string) => {
  const [copied, setCopied] = useState(false);
  const [supported, setSupported] = useState(false);

  useEffect(() => {
    if (!navigator.clipboard) {
      return setSupported(false);
    }
    setSupported(true);
  }, []);

  const copy = useCallback(async () => {
    if (!supported) return;
    try {
      await navigator.clipboard.writeText(copyText);
      setCopied(true);
    } catch (err) {
      console.error('Failed to copy text: ', err);
    }
  }, [supported, copyText]);

  const reset = useCallback(() => {
    setCopied(false);
  }, []);

  return {
    copy,
    supported,
    copied,
    reset,
  };
};

export default useClipboardCopy;
