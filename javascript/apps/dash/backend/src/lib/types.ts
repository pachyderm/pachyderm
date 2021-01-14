export interface Context {
  authToken?: string;
  pachdAddress?: string;
}

export type LinkInputData = {
  source: number;
  target: number;
  error?: boolean;
  active?: boolean;
};
