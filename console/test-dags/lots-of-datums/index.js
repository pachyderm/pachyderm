const fs = require('fs');
const LoremIpsum = require('lorem-ipsum').LoremIpsum;

const lorem = new LoremIpsum({
  sentencesPerParagraph: {
    max: 6,
    min: 3,
  },
  wordsPerSentence: {
    max: 16,
    min: 4,
  },
});

fs.readdirSync('/pfs/data').forEach(async (file) => {
  if (parseFloat(file) < 0.5) {
    throw new Error(lorem.generateSentences(1));
  }

  // Can be tuned to test diffent load times
  await new Promise((resolve) => setTimeout(resolve, 500));

  console.log(file);
  console.log(lorem.generateParagraphs(3));

  fs.writeFileSync(
    '/pfs/out/' + Math.random().toString(),
    Math.random().toString(),
  );
});
