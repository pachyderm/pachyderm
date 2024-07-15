const {exec} = require('child_process');

const NUM_COMMITS = process.argv[2] || 50;
const REPO_NAME = process.argv[3] || 'lots-of-commits';

try {
  for (let i = 0; i < NUM_COMMITS; i++) {
    const rand = Math.trunc(Math.random() * 1000000000).toString();
    exec(
      `pachctl put file ${REPO_NAME}@master:${rand}.png -f ../etc/testing/files/fruit.png`,
      (err, stdout, stderr) => {
        if (err) {
          console.log(err);
          return;
        }
        stdout && console.log(`stdout: ${stdout}`);
        stderr && console.log(`stderr: ${stderr}`);
      },
    );
  }
} catch (err) {
  console.error(err);
}
