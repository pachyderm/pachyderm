#!/usr/bin/env python
"""
Example that does inference on an LSTM networks for amazon review analysis

$ python examples/imdb/auto_inference.py --model_weights imdb.p --vocab_file imdb.vocab 
    --review_files /pfs/reviews --output_dir /pfs/out

"""

from __future__ import print_function
from future import standard_library
standard_library.install_aliases()  # triggers E402, hence noqa below
from builtins import input  # noqa
import numpy as np  # noqa
from neon.backends import gen_backend  # noqa
from neon.initializers import Uniform, GlorotUniform  # noqa
from neon.layers import LSTM, Affine, Dropout, LookupTable, RecurrentSum  # noqa
from neon.models import Model  # noqa
from neon.transforms import Logistic, Tanh, Softmax  # noqa
from neon.util.argparser import NeonArgparser, extract_valid_args  # noqa
from neon.util.compat import pickle  # noqa
from neon.data.text_preprocessing import clean_string  # noqa
import os

# parse the command line arguments
parser = NeonArgparser(__doc__)
parser.add_argument('--model_weights', required=True,
                    help='pickle file of trained weights')
parser.add_argument('--vocab_file', required=True,
                    help='vocabulary file')
parser.add_argument('--review_files', required=True,
                    help='directory containing reviews in text files')
parser.add_argument('--output_dir', required=True,
                    help='directory where results will be saved')
args = parser.parse_args()


# hyperparameters from the reference
batch_size = 1
clip_gradients = True
gradient_limit = 5
vocab_size = 20000
sentence_length = 128
embedding_dim = 128
hidden_size = 128
reset_cells = True
num_epochs = args.epochs

# setup backend
be = gen_backend(**extract_valid_args(args, gen_backend))
be.bsz = 1


# define same model as in train
init_glorot = GlorotUniform()
init_emb = Uniform(low=-0.1 / embedding_dim, high=0.1 / embedding_dim)
nclass = 2
layers = [
    LookupTable(vocab_size=vocab_size, embedding_dim=embedding_dim, init=init_emb,
                pad_idx=0, update=True),
    LSTM(hidden_size, init_glorot, activation=Tanh(),
         gate_activation=Logistic(), reset_cells=True),
    RecurrentSum(),
    Dropout(keep=0.5),
    Affine(nclass, init_glorot, bias=init_glorot, activation=Softmax())
]


# load the weights
print("Initialized the models - ")
model_new = Model(layers=layers)
print("Loading the weights from {0}".format(args.model_weights))

model_new.load_params(args.model_weights)
model_new.initialize(dataset=(sentence_length, batch_size))

# setup buffers before accepting reviews
xdev = be.zeros((sentence_length, 1), dtype=np.int32)  # bsz is 1, feature size
xbuf = np.zeros((1, sentence_length), dtype=np.int32)
oov = 2
start = 1
index_from = 3
pad_char = 0
vocab, rev_vocab = pickle.load(open(args.vocab_file, 'rb'))

# walk over the reviews in the text files, making inferences
for dirpath, dirs, files in os.walk(args.review_files):
    for file in files:
        with open(os.path.join(dirpath, file), 'r') as myfile:
                data=myfile.read()

                # clean the input
                tokens = clean_string(data).strip().split()

                # check for oov and add start
                sent = [len(vocab) + 1 if t not in vocab else vocab[t] for t in tokens]
                sent = [start] + [w + index_from for w in sent]
                sent = [oov if w >= vocab_size else w for w in sent]

                # pad sentences
                xbuf[:] = 0
                trunc = sent[-sentence_length:]
                xbuf[0, -len(trunc):] = trunc
                xdev[:] = xbuf.T.copy()
                y_pred = model_new.fprop(xdev, inference=True)  # inference flag dropout

                with open(os.path.join(args.output_dir, file), "w") as output_file:
                        output_file.write("Pred - {0}\n".format(y_pred.get().T))
