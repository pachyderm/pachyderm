import {Story} from '@pachyderm/components';

import EasyToUnderstandDataLineage from './EasyToUnderstandDataLineage';
import EasyToUseDAGs from './EasyToUseDAGs';
import EasyToUsePipelines from './EasyToUsePipelines';
import EasyToUseVersionedData from './EasyToUseVersionedData';
import IncrementalScalability from './IncrementalScalability';

const stories: Story[] = [
  EasyToUsePipelines,
  EasyToUseVersionedData,
  EasyToUseDAGs,
  EasyToUnderstandDataLineage,
  IncrementalScalability,
];

export default stories;
