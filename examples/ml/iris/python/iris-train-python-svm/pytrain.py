import pandas as pd
from sklearn import svm
from sklearn.externals import joblib
import argparse
import os

# command line arguments
parser = argparse.ArgumentParser(description='Train a model for iris classification.')
parser.add_argument('indir', type=str, help='Input directory containing the training set')
parser.add_argument('outdir', type=str, help='Output directory for the trained model')
args = parser.parse_args()

# training set column names
cols = [
    "Sepal_Length",
    "Sepal_Width",
    "Petal_Length",
    "Petal_Width",
    "Species"
]

features = [
    "Sepal_Length",
    "Sepal_Width",
    "Petal_Length",
    "Petal_Width"
]

# import the iris training set
irisDF = pd.read_csv(os.path.join(args.indir, "iris.csv"), names=cols)

# fit the model
svc = svm.SVC(kernel='linear', C=1.0).fit(irisDF[features], irisDF["Species"])

# output a text description of the model
f = open(os.path.join(args.outdir, 'model.txt'), 'w')
f.write(str(svc))
f.close()

# persist the model
joblib.dump(svc, os.path.join(args.outdir, 'model.pkl'))
