import React from 'react';

export type TaskComponentProps = {
  currentTask: number;
  currentStory: number;
  onCompleted: () => void;
  minimized: boolean;
  index: number;
  name: Section['taskName'];
};

export type Section = {
  header?: React.ReactNode;
  info?: React.ReactNode;
  taskName?: React.ReactNode;
  Task?: React.FC<TaskComponentProps>;
  followUp?: React.ReactNode;
  isSubHeader?: boolean;
};

export type Story = {
  name: string;
  sections: Section[];
};
