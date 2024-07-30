const getListTitle = (noun: string, length: number) => {
  let title = `Last ${length} ${noun}s`;
  if (length === 1) title = `Last ${noun}`;
  if (length === 0) title = `No ${noun}s Found`;

  return title;
};

export default getListTitle;
