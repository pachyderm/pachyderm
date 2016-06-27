import os

input_dir = '/pfs/wordcount_map'
output_dir = '/pfs/wordcount_reduce'

for filename in os.listdir(input_dir):
    with open(os.path.join(input_dir, filename), 'r') as f:
        sum = 0
        for line in f:
            sum += int(line)
        with open(os.path.join(output_dir, filename), 'w') as f:
            f.write(str(sum))
