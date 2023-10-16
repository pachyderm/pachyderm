def test_error(t):
    t.Error("this is an error", "this too")
    t.Errorf("%v", "this is also an error")

def test_fatal(t):
    print("start")
    t.Fatal("this has failed", "severely")
    t.Error("you should not see this")
    print("you should also not see this")

def test_unexpected_kwargs(t):
    t.Errorf(message = "foo")

def test_invalid_parameters():
    pass

def test_extra_parameters(foo, bar, baz = "quux"):
    pass

test_variable = "this is not a test"

def this_is_not_a_test(t):
    t.Fatal("this should not have run")

print("this is logged at the top level")

def test_bad_calls(t):
    print()
    t.Error()
    t.Logf()
    t.Errorf(42, 43)
