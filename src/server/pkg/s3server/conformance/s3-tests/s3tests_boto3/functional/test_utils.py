from nose.tools import eq_ as eq

import utils

def test_generate():
    FIVE_MB = 5 * 1024 * 1024
    eq(len(''.join(utils.generate_random(0))), 0)
    eq(len(''.join(utils.generate_random(1))), 1)
    eq(len(''.join(utils.generate_random(FIVE_MB - 1))), FIVE_MB - 1)
    eq(len(''.join(utils.generate_random(FIVE_MB))), FIVE_MB)
    eq(len(''.join(utils.generate_random(FIVE_MB + 1))), FIVE_MB + 1)
