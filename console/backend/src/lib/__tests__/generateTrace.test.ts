import {generateConsoleTraceUuid} from '../generateTrace';

describe('generateConsoleTraceUuid', () => {
  // Test 10 so as to eliminate a possible flake without having to use a deterministic uuid
  const items = Array.from(Array(10)).map(() => generateConsoleTraceUuid());
  test.each(items)('generates a uuid with 3 as the 15th nibble', (value) => {
    const uuidRegexWith15thNibbleAs3 =
      /^[0-9a-fA-F]{8}\b-[0-9a-fA-F]{4}\b-3[0-9a-fA-F]{3}\b-[0-9a-fA-F]{4}\b-[0-9a-fA-F]{12}$/;
    expect(value).toMatch(uuidRegexWith15thNibbleAs3);
  });
});
