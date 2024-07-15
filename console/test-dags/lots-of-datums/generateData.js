const fs = require('fs');
const path = require('path');

const NUM_DATUMS = process.argv[2] || 50;
const writeDir = process.argv[3] || './data';

fs.readdir(writeDir, (err, files) => {
  if (err) throw err;

  for (const file of files) {
    fs.unlink(path.join(writeDir, file), (err) => {
      if (err) throw err;
    });
  }
});

try {
  for (let i = 0; i < NUM_DATUMS; i++) {
    const rand = Math.random().toString();
    fs.writeFileSync(writeDir + '/' + rand, rand);
  }
} catch (err) {
  console.error(err);
}
