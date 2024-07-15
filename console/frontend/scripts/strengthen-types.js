/* eslint-disable @typescript-eslint/no-var-requires */

const {Project, SyntaxKind} = require('ts-morph');
const {SemicolonPreference} = require('typescript');

const addTypeNamesToTypes = (
  sourceFiles = './src/generated/proto/**/*.pb.ts',
) => {
  const project = new Project({});
  let updatedTypesCount = 0;

  project.addSourceFilesAtPaths(sourceFiles);
  project.getSourceFiles().forEach((sourceFile) => {
    sourceFile.getDescendantsOfKind(SyntaxKind.TypeLiteral).forEach((des) => {
      const parent = des.getParentIfKind(SyntaxKind.TypeAliasDeclaration);

      // Only add properties to top level types
      if (!parent) return;

      // Don't add a __typename if it already exists
      if (des.getProperty('__typename')) return;

      const typeName = parent.getName();

      // Don't add a __typename if there is no name
      if (!typeName) return;

      des.insertProperty(0, {
        name: '__typename',
        type: `"${typeName}"`,
        hasQuestionToken: true,
      });

      updatedTypesCount += 1;
    });

    sourceFile.formatText({
      semicolons: SemicolonPreference.Insert,
      tabSize: 2,
      convertTabsToSpaces: true,
    });
  });

  project.saveSync();

  console.log(`Updated ${updatedTypesCount} types`);
};

module.exports = {
  addTypeNamesToTypes,
};
