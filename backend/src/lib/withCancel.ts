const withCancel = <T>(
  asyncIterator: AsyncIterator<T>,
  onCancel: () => void,
): AsyncIterator<T> => {
  if (!asyncIterator.return) {
    asyncIterator.return = () =>
      Promise.resolve({value: undefined, done: true});
  }

  const savedReturn = asyncIterator.return.bind(asyncIterator);
  asyncIterator.return = () => {
    onCancel();
    return savedReturn();
  };

  return asyncIterator;
};

export default withCancel;
