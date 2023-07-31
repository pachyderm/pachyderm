export const generateUUID = () => {
  const allowed_characters = '0123456789abcdef';
  let id = '';

  for (let i = 0; i < 32; i++) {
    if (i === 12) {
      id += '4';
    } else {
      const randomIndex = Math.floor(Math.random() * allowed_characters.length);
      id += allowed_characters[randomIndex];
    }
  }
  return id;
};

export default generateUUID;
