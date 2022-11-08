import useDownloadText from '@dash-frontend/views/LogsViewers/LogsViewer/hooks/useDownloadText';
import {useClipboardCopy} from '@pachyderm/components';

import stringifyToFormat, {Format} from './utils/stringifyToFormat';

interface useConfigFilePreviewProps {
  config: Record<string, unknown> | [];
}

const useConfigFilePreview = ({config}: useConfigFilePreviewProps) => {
  const {copy} = useClipboardCopy(stringifyToFormat(config, Format.YAML));
  const {download: downloadJSON} = useDownloadText(
    stringifyToFormat(config, Format.JSON),
    `job_definition.json`,
  );
  const {download: downloadYAML} = useDownloadText(
    stringifyToFormat(config, Format.YAML),
    `job_definition.yaml`,
  );

  const onSelectGearAction = (action: string) => {
    if (action === 'copy') {
      copy();
    }
    if (action === 'download-json') {
      downloadJSON();
    }
    if (action === 'download-yaml') {
      downloadYAML();
    }
  };

  return {
    onSelectGearAction,
  };
};

export default useConfigFilePreview;
