import {useEffect, useState} from 'react';

import {ViewState} from '@dash-frontend/hooks/useUrlQueryState';
import downloadSVG from '@dash-frontend/lib/downloadSVG';

export const useCanvasDownload = (
  searchParams: ViewState,
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
          searchParams.globalIdFilter,
        );
      download().then(() => setCanvasDownload(false));
    }
  }, [
    graphExtents.xMax,
    graphExtents.yMax,
    projectName,
    renderAndDownloadCanvas,
    searchParams.globalIdFilter,
  ]);

  return {
    renderAndDownloadCanvas,
    downloadCanvas,
  };
};
