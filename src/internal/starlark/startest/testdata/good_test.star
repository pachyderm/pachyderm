def test_two_plus_two(t):
    got, want = 2 + 2, 4
    print("got, want = %d, %d" % (got, want))
    if got != want:
        t.Errorf("2 + 2 is not equal to 4:\n  got: %v\n want: %v", got, want)
