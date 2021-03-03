export const isTest = () => {
  return process.env.JEST_WORKER_ID !== undefined;
};
