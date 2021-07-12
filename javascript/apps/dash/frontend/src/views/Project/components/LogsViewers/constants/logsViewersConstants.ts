import {DropdownItem} from '@pachyderm/components';

export const DEFAULT_ROW_HEIGHT = 46;
export const HEADER_HEIGHT_OFFSET = 153;
export const LOGS_DEFAULT_DROPDOWN_OPTIONS: DropdownItem[] = [
  {id: 'Last 30 Minutes', content: 'Last 30 Minutes', closeOnClick: true},
  {id: 'Last 24 Hours', content: 'Last 24 Hours', closeOnClick: true},
  {id: 'Last 3 Days', content: 'Last 3 Days', closeOnClick: true},
];
export const PIPELINE_START_TIME_OPTION = 'Last Pipeline Job';
export const JOB_START_TIME_OPTION = 'Job Start Time';

export const LOGS_DATE_FORMAT = 'MM-dd-yyyy H:mm:ss';
