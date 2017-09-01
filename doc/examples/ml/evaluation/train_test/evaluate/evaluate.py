import pandas as pd
from sklearn.externals import joblib
from sklearn.metrics import accuracy_score
import argparse
import os
import re

# command line arguments
parser = argparse.ArgumentParser(description='Evaluate a model for iris classification.')
parser.add_argument('model', type=str, help='Input directory containing the training set')
parser.add_argument('testdata', type=str, help='Input directory containing the test set')
parser.add_argument('out', type=str, help='Output directory for the evaluation results')
args = parser.parse_args()

# attribute column names
features = [
    "Sepal_Length",
    "Sepal_Width",
    "Petal_Length",
    "Petal_Width"
]

response = ["Species"]

# load the model
mymodel = joblib.load(args.model) 

# read in the test data
testData = pd.read_csv(args.testdata, names=features+response)

# make the inference
predictions = mymodel.predict(testData[features])

# calculate the accuracy
accuracy = accuracy_score(testData[response], predictions)

# save the accuracy
model_name = re.sub('\.pkl$', '', args.model.split("/")[-1])
text_file = open(os.path.join(args.out, model_name + "_metric.txt"), "w")
text_file.write(str(accuracy))
text_file.close()  
