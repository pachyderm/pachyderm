import {createEffectiveSpecDecorationMap} from '../createEffectiveSpecDecorationMap';

describe('createEffectiveSpecDecorationMap', () => {
  it('real data example', () => {
    expect(
      createEffectiveSpecDecorationMap(
        {
          pipeline: {
            name: 'goofy-donkey',
          },
          transform: {cmd: ['python']},
          description: 'wonderful new description',
        },
        {
          description: 'old',
          resourceRequests: {
            cpu: 1,
          },
        },
      ),
    ).toEqual({
      pipeline: {name: null},
      transform: {cmd: null},
      description: 'old',
    });
  });

  it('assigns the previous value from defaults to leaves when the value differs', () => {
    expect(
      createEffectiveSpecDecorationMap(
        {
          key: 'new',
          object: {key: 'new'},
        },
        {key: 'old', object: {key: 'old'}},
      ),
    ).toEqual({
      key: 'old',
      object: {key: 'old'},
    });
  });

  it('assigns null to arrays', () => {
    expect(
      createEffectiveSpecDecorationMap(
        {
          array: ['foo'],
          object: {array: ['bar']},
          arrayObject: [{foo: 'baz'}],
        },
        {},
      ),
    ).toEqual({
      array: null,
      object: {array: null},
      arrayObject: null,
    });
  });

  it('assigns null to values not present in createPipelineDefaults', () => {
    expect(
      createEffectiveSpecDecorationMap(
        {
          netNewValue: true,
        },
        {},
      ),
    ).toEqual({
      netNewValue: null,
    });
  });

  it('assigns null to values that match in the user spec and defaults', () => {
    expect(
      createEffectiveSpecDecorationMap(
        {
          matching: 'value',
        },
        {matching: 'value'},
      ),
    ).toEqual({
      matching: null,
    });
  });

  it('values present in default but not user spec do not appear', () => {
    expect(
      createEffectiveSpecDecorationMap({}, {shouldNotAppear: true}),
    ).toEqual({});
  });
});
