import {useCallback, useState, useEffect} from 'react';

const useClipboardCopy = (copyText: string) => {
  const [copied, setCopied] = useState(false);
  const [supported, setSupported] = useState(false);

  useEffect(() => {
    if (
      typeof document !== 'undefined' &&
      document.queryCommandSupported('copy')
    ) {
      setSupported(true);
    }
  }, [setSupported]);

  const copy = useCallback(() => {
    if (!supported) return;

    const area = document.createElement('textarea');
    area.style.cssText =
      'position: absolute; left: -999em; top: -999em, opacity: 0';
    area.setAttribute('aria-hidden', 'true');
    area.value = copyText;
    document.body.appendChild(area);
    area.select();
    document.execCommand('copy');
    area.remove();
    setCopied(true);
  }, [supported, copyText, setCopied]);

  const reset = useCallback(() => {
    setCopied(false);
  }, [setCopied]);

  return {
    copy,
    supported,
    copied,
    reset,
  };
};

export default useClipboardCopy;
