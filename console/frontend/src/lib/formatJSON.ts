const formatJSON = (jsonString: string) => {
  return JSON.stringify(JSON.parse(jsonString), null, 2);
};

export default formatJSON;
