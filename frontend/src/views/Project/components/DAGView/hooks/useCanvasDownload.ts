import {useEffect, useState} from 'react';

import {UrlState} from '@dash-frontend/hooks/useUrlQueryState';
import downloadSVG from '@dash-frontend/lib/downloadSVG';

export const useCanvasDownload = (
  viewState: UrlState,
  graphExtents: {
    xMin: number;
    xMax: number;
    yMin: number;
    yMax: number;
  },
  projectName?: string,
) => {
  const [renderAndDownloadCanvas, setCanvasDownload] = useState(false);

  const downloadCanvas = () => {
    setCanvasDownload(true);
  };

  useEffect(() => {
    if (renderAndDownloadCanvas) {
      const download = async () =>
        await downloadSVG(
          graphExtents.xMax + 100,
          graphExtents.yMax + 100,
          projectName,
          viewState.globalIdFilter,
        );
      download().then(() => setCanvasDownload(false));
    }
  }, [
    graphExtents.xMax,
    graphExtents.yMax,
    projectName,
    renderAndDownloadCanvas,
    viewState.globalIdFilter,
  ]);

  return {
    renderAndDownloadCanvas,
    downloadCanvas,
  };
};
