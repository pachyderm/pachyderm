from s3tests import realistic
import shutil
import tempfile


# XXX not used for now
def create_files(mean=2000):
    return realistic.files2(
        mean=1024 * mean,
        stddev=1024 * 500,
        seed=1256193726,
        numfiles=4,
    )


class TestFiles(object):
    # the size and seed is what we can get when generating a bunch of files
    # with pseudo random numbers based on sttdev, seed, and mean.

    # this fails, demonstrating the (current) problem
    #def test_random_file_invalid(self):
    #    size = 2506764
    #    seed = 3391518755
    #    source = realistic.RandomContentFile(size=size, seed=seed)
    #    t = tempfile.SpooledTemporaryFile()
    #    shutil.copyfileobj(source, t)
    #    precomputed = realistic.PrecomputedContentFile(t)
    #    assert precomputed.valid()

    #    verifier = realistic.FileVerifier()
    #    shutil.copyfileobj(precomputed, verifier)

    #    assert verifier.valid()

    # this passes
    def test_random_file_valid(self):
        size = 2506001
        seed = 3391518755
        source = realistic.RandomContentFile(size=size, seed=seed)
        t = tempfile.SpooledTemporaryFile()
        shutil.copyfileobj(source, t)
        precomputed = realistic.PrecomputedContentFile(t)

        verifier = realistic.FileVerifier()
        shutil.copyfileobj(precomputed, verifier)

        assert verifier.valid()


# new implementation
class TestFileValidator(object):

    def test_new_file_is_valid(self):
        size = 2506001
        contents = realistic.generate_file_contents(size)
        t = tempfile.SpooledTemporaryFile()
        t.write(contents)
        t.seek(0)
        fp = realistic.FileValidator(t)
        assert fp.valid()

    def test_new_file_is_valid_when_size_is_1(self):
        size = 1
        contents = realistic.generate_file_contents(size)
        t = tempfile.SpooledTemporaryFile()
        t.write(contents)
        t.seek(0)
        fp = realistic.FileValidator(t)
        assert fp.valid()

    def test_new_file_is_valid_on_several_calls(self):
        size = 2506001
        contents = realistic.generate_file_contents(size)
        t = tempfile.SpooledTemporaryFile()
        t.write(contents)
        t.seek(0)
        fp = realistic.FileValidator(t)
        assert fp.valid()
        assert fp.valid()
