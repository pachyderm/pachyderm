export enum MenuActions {
  Close,
  CloseSelect,
  First,
  Last,
  Next,
  Open,
  PageDown,
  PageUp,
  Previous,
  Select,
  Space,
}

export interface OptionRef {
  value: string;
  displayValue: React.ReactNode;
}
