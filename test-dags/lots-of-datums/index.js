const fs = require("fs");

fs.readdirSync("/pfs/data").forEach(async (file) => {
  // Can be tuned to test diffent load times
  await new Promise((resolve) => setTimeout(resolve, 500));

  console.log(file);
  fs.writeFileSync(
    "/pfs/out/" + Math.random().toString(),
    Math.random().toString()
  );
});
