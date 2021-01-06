export type LinkDatum = {
  source: Coordinate;
  target: Coordinate;
  error?: boolean;
  active?: boolean;
};

export type NodeDatum = {
  name: string | number | boolean;
  error?: boolean;
  type?: string;
  success?: boolean;
  access?: boolean;
  x?: number;
  y?: number;
  fx?: number;
  fy?: number;
};

export type Coordinate = {
  x: number;
  y: number;
};

export type LinkInputData = {
  source: number;
  target: number;
  error?: boolean;
  active?: boolean;
};

export type DataProps = {
  links: LinkInputData[];
  nodes: NodeDatum[];
};
