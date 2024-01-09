import capitalize from 'lodash/capitalize';

import {PipelineInfoPipelineType} from '@dash-frontend/api/pps';

const readablePipelineType = (PipelineType?: PipelineInfoPipelineType) => {
  if (!PipelineType) return '';
  switch (PipelineType) {
    case PipelineInfoPipelineType.PIPELINE_TYPE_SERVICE: {
      return capitalize('Service');
    }
    case PipelineInfoPipelineType.PIPELINE_TYPE_SPOUT: {
      return capitalize('Spout');
    }
    case PipelineInfoPipelineType.PIPELINE_TYPE_TRANSFORM: {
      return capitalize('Standard');
    }
    case PipelineInfoPipelineType.PIPELINT_TYPE_UNKNOWN: {
      return capitalize('Unknown');
    }
    default:
      return '';
  }
};

export default readablePipelineType;
