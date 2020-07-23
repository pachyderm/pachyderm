import argparse
import os
from os import path
import numpy as np
import pandas as pd
from sklearn.model_selection import ShuffleSplit
import matplotlib.pyplot as plt
import seaborn as sns
import joblib
from utils import plot_learning_curve

from sklearn import datasets, ensemble, linear_model
from sklearn.model_selection import learning_curve
from sklearn.model_selection import ShuffleSplit
from sklearn.model_selection import GridSearchCV

parser = argparse.ArgumentParser(description="Structured data regression")
parser.add_argument("--input",
                    type=str,
                    help="csv file with all examples")
parser.add_argument("--target-col",
                    type=str,
                    help="column with target values")
parser.add_argument("--output",
                    metavar="DIR",
                    default='./output',
                    help="output directory")

def load_data(input_csv, target_col):
    # Load the Boston housing dataset
    data = pd.read_csv(input_csv, header=0)
    targets = data[target_col]
    features = data.drop(target_col, axis = 1)
    print("Dataset has {} data points with {} variables each.".format(*data.shape))
    return data, features, targets

def create_pairplot(data):
    plt.clf()
    # Calculate and show pairplot
    sns.pairplot(data, height=2.5)
    plt.tight_layout()

def create_corr_matrix(data):
    plt.clf()
    # Calculate and show correlation matrix
    sns.set()
    corr = data.corr()
    # Generate a mask for the upper triangle
    mask = np.triu(np.ones_like(corr, dtype=np.bool))

    # Generate a custom diverging colormap
    cmap = sns.diverging_palette(220, 10, as_cmap=True)

    # Draw the heatmap with the mask and correct aspect ratio
    sns_plot = sns.heatmap(corr, mask=mask, cmap=cmap, vmax=.3, center=0,
                square=True, linewidths=.5, cbar_kws={"shrink": .5})

def train_model(estimator, features, targets):
    parameters = {
        "max_depth":[3,5,8],
        "criterion": ["friedman_mse",  "mae"],
        "subsample":[0.5, 0.618, 0.8, 0.85, 0.9, 0.95, 1.0],
        "n_estimators":[10]
    }

    clf = GridSearchCV(estimator, parameters, cv=10, n_jobs=4)
    return clf

def create_learning_curve(estimator, features, targets):
    plt.clf()
    fig, axes = plt.subplots(3, 1, figsize=(5, 5))

    title = "Learning Curves (Gradient Boosting Regressor)"
    # Gradient Boosting is more expensive so we do a lower number of CV iterations:
    cv = ShuffleSplit(n_splits=10, test_size=0.2, random_state=0)
    # estimator = ensemble.GradientBoostingRegressor()
    plot_learning_curve(estimator, title, features, targets, axes=axes[:], 
                        ylim=(0.7, 1.01), cv=cv, n_jobs=4)

def main():
    args = parser.parse_args()
    os.makedirs(args.output, exist_ok=True)
    
    # Data loading and Exploration
    data, features, targets = load_data(args.input, args.target_col)
    create_pairplot(data)
    plt.savefig(path.join(args.output, 'pairplot.png'))
    create_corr_matrix(data)
    plt.savefig(path.join(args.output, 'corr_matrix.png'))

    # Fit model
    estimator = train_model(ensemble.GradientBoostingRegressor(), features, targets)
    create_learning_curve(estimator, features, targets)
    plt.savefig(path.join(args.output, 'cv_reg_output.png'))

    # Save model
    joblib.dump(estimator, path.join(args.output, 'final_model.sav'))

if __name__ == "__main__":
    main()