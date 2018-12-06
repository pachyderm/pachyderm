import pandas as pd
from sklearn.externals import joblib
import argparse
import os

# command line arguments
parser = argparse.ArgumentParser(description='Train a model for iris classification.')
parser.add_argument('inmodeldir', type=str, help='Input directory containing the training set')
parser.add_argument('inattdir', type=str, help='Input directory containing the input attributes')
parser.add_argument('outdir', type=str, help='Output directory for the trained model')
args = parser.parse_args()

# attribute column names
features = [
    "Sepal_Length",
    "Sepal_Width",
    "Petal_Length",
    "Petal_Width"
]

# load the model
mymodel = joblib.load(os.path.join(args.inmodeldir, 'model.pkl')) 

# walk the input attributes directory and make an
# inference for every attributes file found
for dirpath, dirs, files in os.walk(args.inattdir):
    for file in files:

        # read in the attributes
        attr = pd.read_csv(os.path.join(dirpath, file), names=features)

        # make the inference
        pred = mymodel.predict(attr)

        # save the inference
        output = pd.DataFrame(pred, columns=["Species"])
        output.to_csv(os.path.join(args.outdir, file.split(".")[0]), header=False, index=False)

