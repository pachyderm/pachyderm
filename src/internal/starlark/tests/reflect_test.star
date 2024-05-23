def test_struct(t):
    if struct.Field != "simple":
        t.Errorf("struct.Field: got %v, want %v", struct.Field, "simple")
    if struct["Field"] != "simple":
        t.Errorf("struct.Field: got %v, want %v", struct["Field"], "simple")
    if struct_ptr != struct.Me:
        t.Errorf("struct.Me: got %v, want %v", struct.Me, struct_ptr)
    if type(struct) != "starlark_test.SimpleStruct":
        t.Errorf("type(struct): got %v, want %v", type(struct), "starlark_test.SimpleStruct")

    want_fields = ["Field", "Me"]
    if dir(struct) != want_fields:
        t.Errorf("dir(struct): got %v, want %v", dir(struct), want_fields)

def test_struct_pointer(t):
    if struct_ptr.Field != "simple":
        t.Errorf("struct_ptr.Field: got %v, want %v", struct_ptr.Field, "simple")
    if struct_ptr["Field"] != "simple":
        t.Errorf("struct_ptr.Field: got %v, want %v", struct_ptr["Field"], "simple")
    if struct_ptr != struct_ptr.Me:
        t.Errorf("struct_ptr.Me: got %v, want %v", struct_ptr.Me, struct_ptr)
    if str(struct_ptr) != "simple struct":
        t.Errorf("str(struct_ptr): got %v, want %v", string(struct_ptr), "simple struct")
    if type(struct_ptr) != "*starlark_test.SimpleStruct":
        t.Errorf("type(struct_ptr): got %v, want %v", type(struct_ptr), "*starlark_test.SimpleStruct")

    want_fields = ["Field", "Me"]
    if dir(struct_ptr) != want_fields:
        t.Errorf("dir(struct_ptr): got %v, want %v", dir(struct_ptr), want_fields)

def test_basic_types(t):
    gotdata = {
        "string": string,
        "bool": bool,
        "bytes": byte,
        "int": int,
        "float": float,
        "stringlist": stringlist,
        "array": array,
        "int8": int8,
        "uint8": uint8,
    }
    wantdata = {
        "string": "string",
        "bool": True,
        "bytes": bytes("bytes"),
        "int": 42,
        "float": 123.45,
        "stringlist": ["string", "string"],
        "array": [0, 0, 0, 0, 0, 5, 0, 0, 0, 0],
        "int8": -128,
        "uint8": 255,
    }
    for k, want in wantdata.items():
        got = gotdata[k]
        if got != want:
            t.Errorf("item %v:\n  got: %v\n want: %v", k, got, want)

    wantdata = {"one": 1, "two": 2}
    for k, want in wantdata.items():
        got = map[k]
        if got != want:
            t.Errorf("item %v:\n  got: %v\n want: %v", k, got, want)

def test_wrapped_nil(t):
    print(str(wrappednil))

def test_putting_structs_into_a_map(t):
    x = {struct: 42, struct_ptr: 43}
