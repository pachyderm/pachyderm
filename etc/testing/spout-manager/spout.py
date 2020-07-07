import python_pachyderm
import random
import string
import time

for i in range(10):
    with python_pachyderm.SpoutManager() as spout:
        content = ''.join(random.choice(string.ascii_lowercase) for j in range(2048))
        spout.add_from_bytes(str(i), content.encode())
    time.sleep(5)
