import {useCallback, useState} from 'react';

const useDownloadText = (downloadText: string, fileName?: string) => {
  const [downloaded, setDownloaded] = useState(false);

  const download = useCallback(() => {
    const element = document.createElement('a');
    element.setAttribute(
      'href',
      'data:text/plain;charset=utf-8,' + encodeURIComponent(downloadText),
    );
    element.setAttribute('download', fileName || 'data');

    element.style.display = 'none';
    document.body.appendChild(element);
    element.click();
    document.body.removeChild(element);
    setDownloaded(true);
  }, [downloadText, fileName]);

  const reset = useCallback(() => {
    setDownloaded(false);
  }, [setDownloaded]);

  return {download, downloaded, reset};
};
export default useDownloadText;
