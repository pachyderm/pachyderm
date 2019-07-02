import hashlib
import random
import string
import struct
import time
import math
import tempfile
import shutil
import os


NANOSECOND = int(1e9)


def generate_file_contents(size):
    """
    A helper function to generate binary contents for a given size, and
    calculates the md5 hash of the contents appending itself at the end of the
    blob.
    It uses sha1's hexdigest which is 40 chars long. So any binary generated
    should remove the last 40 chars from the blob to retrieve the original hash
    and binary so that validity can be proved.
    """
    size = int(size)
    contents = os.urandom(size)
    content_hash = hashlib.sha1(contents).hexdigest()
    return contents + content_hash


class FileValidator(object):

    def __init__(self, f=None):
        self._file = tempfile.SpooledTemporaryFile()
        self.original_hash = None
        self.new_hash = None
        if f:
            f.seek(0)
            shutil.copyfileobj(f, self._file)

    def valid(self):
        """
        Returns True if this file looks valid. The file is valid if the end
        of the file has the md5 digest for the first part of the file.
        """
        self._file.seek(0)
        contents = self._file.read()
        self.original_hash, binary = contents[-40:], contents[:-40]
        self.new_hash = hashlib.sha1(binary).hexdigest()
        if not self.new_hash == self.original_hash:
            print 'original  hash: ', self.original_hash
            print 'new hash: ', self.new_hash
            print 'size: ', self._file.tell()
            return False
        return True

    # XXX not sure if we need all of these
    def seek(self, offset, whence=os.SEEK_SET):
        self._file.seek(offset, whence)

    def tell(self):
        return self._file.tell()

    def read(self, size=-1):
        return self._file.read(size)

    def write(self, data):
        self._file.write(data)
        self._file.seek(0)


class RandomContentFile(object):
    def __init__(self, size, seed):
        self.size = size
        self.seed = seed
        self.random = random.Random(self.seed)

        # Boto likes to seek once more after it's done reading, so we need to save the last chunks/seek value.
        self.last_chunks = self.chunks = None
        self.last_seek = None

        # Let seek initialize the rest of it, rather than dup code
        self.seek(0)

    def _mark_chunk(self):
        self.chunks.append([self.offset, int(round((time.time() - self.last_seek) * NANOSECOND))])

    def seek(self, offset, whence=os.SEEK_SET):
        if whence == os.SEEK_SET:
            self.offset = offset
        elif whence == os.SEEK_END:
            self.offset = self.size + offset;
        elif whence == os.SEEK_CUR:
            self.offset += offset

        assert self.offset == 0

        self.random.seed(self.seed)
        self.buffer = ''

        self.hash = hashlib.md5()
        self.digest_size = self.hash.digest_size
        self.digest = None

        # Save the last seek time as our start time, and the last chunks
        self.last_chunks = self.chunks
        # Before emptying.
        self.last_seek = time.time()
        self.chunks = []

    def tell(self):
        return self.offset

    def _generate(self):
        # generate and return a chunk of pseudorandom data
        size = min(self.size, 1*1024*1024) # generate at most 1 MB at a time
        chunks = int(math.ceil(size/8.0))  # number of 8-byte chunks to create

        l = [self.random.getrandbits(64) for _ in xrange(chunks)]
        s = struct.pack(chunks*'Q', *l)
        return s

    def read(self, size=-1):
        if size < 0:
            size = self.size - self.offset

        r = []

        random_count = min(size, self.size - self.offset - self.digest_size)
        if random_count > 0:
            while len(self.buffer) < random_count:
                self.buffer += self._generate()
            self.offset += random_count
            size -= random_count
            data, self.buffer = self.buffer[:random_count], self.buffer[random_count:]
            if self.hash is not None:
                self.hash.update(data)
            r.append(data)

        digest_count = min(size, self.size - self.offset)
        if digest_count > 0:
            if self.digest is None:
                self.digest = self.hash.digest()
                self.hash = None
            self.offset += digest_count
            size -= digest_count
            data = self.digest[:digest_count]
            r.append(data)

        self._mark_chunk()

        return ''.join(r)


class PrecomputedContentFile(object):
    def __init__(self, f):
        self._file = tempfile.SpooledTemporaryFile()
        f.seek(0)
        shutil.copyfileobj(f, self._file)

        self.last_chunks = self.chunks = None
        self.seek(0)

    def seek(self, offset, whence=os.SEEK_SET):
        self._file.seek(offset, whence)

        if self.tell() == 0:
            # only reset the chunks when seeking to the beginning
            self.last_chunks = self.chunks
            self.last_seek = time.time()
            self.chunks = []

    def tell(self):
        return self._file.tell()

    def read(self, size=-1):
        data = self._file.read(size)
        self._mark_chunk()
        return data

    def _mark_chunk(self):
        elapsed = time.time() - self.last_seek
        elapsed_nsec = int(round(elapsed * NANOSECOND))
        self.chunks.append([self.tell(), elapsed_nsec])

class FileVerifier(object):
    def __init__(self):
        self.size = 0
        self.hash = hashlib.md5()
        self.buf = ''
        self.created_at = time.time()
        self.chunks = []

    def _mark_chunk(self):
        self.chunks.append([self.size, int(round((time.time() - self.created_at) * NANOSECOND))])

    def write(self, data):
        self.size += len(data)
        self.buf += data
        digsz = -1*self.hash.digest_size
        new_data, self.buf = self.buf[0:digsz], self.buf[digsz:]
        self.hash.update(new_data)
        self._mark_chunk()

    def valid(self):
        """
        Returns True if this file looks valid. The file is valid if the end
        of the file has the md5 digest for the first part of the file.
        """
        if self.size < self.hash.digest_size:
            return self.hash.digest().startswith(self.buf)

        return self.buf == self.hash.digest()


def files(mean, stddev, seed=None):
    """
    Yields file-like objects with effectively random contents, where
    the size of each file follows the normal distribution with `mean`
    and `stddev`.

    Beware, the file-likeness is very shallow. You can use boto's
    `key.set_contents_from_file` to send these to S3, but they are not
    full file objects.

    The last 128 bits are the MD5 digest of the previous bytes, for
    verifying round-trip data integrity. For example, if you
    re-download the object and place the contents into a file called
    ``foo``, the following should print two identical lines:

      python -c 'import sys, hashlib; data=sys.stdin.read(); print hashlib.md5(data[:-16]).hexdigest(); print "".join("%02x" % ord(c) for c in data[-16:])' <foo

    Except for objects shorter than 16 bytes, where the second line
    will be proportionally shorter.
    """
    rand = random.Random(seed)
    while True:
        while True:
            size = int(rand.normalvariate(mean, stddev))
            if size >= 0:
                break
        yield RandomContentFile(size=size, seed=rand.getrandbits(32))


def files2(mean, stddev, seed=None, numfiles=10):
    """
    Yields file objects with effectively random contents, where the
    size of each file follows the normal distribution with `mean` and
    `stddev`.

    Rather than continuously generating new files, this pre-computes and
    stores `numfiles` files and yields them in a loop.
    """
    # pre-compute all the files (and save with TemporaryFiles)
    fs = []
    for _ in xrange(numfiles):
        t = tempfile.SpooledTemporaryFile()
        t.write(generate_file_contents(random.normalvariate(mean, stddev)))
        t.seek(0)
        fs.append(t)

    while True:
        for f in fs:
            yield f


def names(mean, stddev, charset=None, seed=None):
    """
    Yields strings that are somewhat plausible as file names, where
    the lenght of each filename follows the normal distribution with
    `mean` and `stddev`.
    """
    if charset is None:
        charset = string.ascii_lowercase
    rand = random.Random(seed)
    while True:
        while True:
            length = int(rand.normalvariate(mean, stddev))
            if length > 0:
                break
        name = ''.join(rand.choice(charset) for _ in xrange(length))
        yield name
