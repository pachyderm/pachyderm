import {DropdownItem} from '@pachyderm/components';

export const DEFAULT_ROW_HEIGHT = 24;
export const HEADER_HEIGHT_OFFSET = 40;

export const LOGS_DEFAULT_DROPDOWN_OPTIONS: DropdownItem[] = [
  {id: 'Last 30 Minutes', content: 'Last 30 Minutes'},
  {id: 'Last 24 Hours', content: 'Last 24 Hours'},
  {id: 'Last 3 Days', content: 'Last 3 Days'},
];

export const LOGS_PAGE_SIZE = 1000;
