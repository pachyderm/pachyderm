import pandas as pd
from sklearn import svm
from sklearn.externals import joblib
import argparse
import os

# command line arguments
parser = argparse.ArgumentParser(description='Train a model for iris classification.')
parser.add_argument('indir', type=str, help='Input directory containing the training set')
parser.add_argument('outdir', type=str, help='Output directory for the trained model')
parser.add_argument('cparam', type=float, help='Parameter C for SVM')
parser.add_argument('gammaparam', type=float, help='Parameter Gamma for SVM')
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
svc = svm.SVC(kernel='linear', C=args.cparam, gamma=args.gammaparam).fit(irisDF[features], irisDF["Species"])

# persist the model
joblib.dump(svc, os.path.join(args.outdir, 'model_C' + str(args.cparam) + '_G' + str(args.gammaparam) + '.pkl'))
