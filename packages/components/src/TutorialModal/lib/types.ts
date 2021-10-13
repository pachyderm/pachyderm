export type TaskComponentProps = {
  currentTask: number;
  onCompleted: () => void;
  minimized: boolean;
  index: number;
  name: Task['name'];
};

export type Task = {
  name: React.ReactNode;
  info?: {
    name: string;
    text: React.ReactNode[];
  };
  Task: React.FC<TaskComponentProps>;
  followUp?: React.ReactNode;
};

export type Step = {
  label?: string;
  name: string;
  tasks: Task[];
  instructionsHeader: string;
  instructionsText: React.ReactNode;
  prerequisites?: React.ReactNode;
};
