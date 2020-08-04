import argparse
import os
from os import path
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import joblib
from utils import plot_learning_curve

from sklearn.model_selection import ShuffleSplit
from sklearn import datasets, ensemble, linear_model
from sklearn.model_selection import learning_curve
from sklearn.model_selection import ShuffleSplit
from sklearn.model_selection import cross_val_score

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
                square=True, linewidths=.5, annot=True, cbar_kws={"shrink": .5})

def train_model(features, targets):
    # Train a Random Forest Regression model
    reg = ensemble.RandomForestRegressor(random_state=1)
    scores = cross_val_score(reg, features, targets, cv=10)
    print("Score: {:2f} (+/- {:2f})".format(scores.mean(), scores.std() * 2))
    return reg

def create_learning_curve(estimator, features, targets):
    plt.clf()

    title = "Learning Curves (Random Forest Regressor)"
    cv = ShuffleSplit(n_splits=10, test_size=0.2, random_state=0)
    plot_learning_curve(estimator, title, features, targets, 
                        ylim=(0.5, 1.01), cv=cv, n_jobs=4)

def main():
    args = parser.parse_args()
    if os.path.isfile(args.input):
        input_files = [args.input]
    else:  # Directory
        for dirpath, dirs, files in os.walk(args.input):  
            input_files = [ os.path.join(dirpath, filename) for filename in files if filename.endswith('.csv') ]
    print("Datasets: {}".format(input_files))
    os.makedirs(args.output, exist_ok=True)

    for filename in input_files:

        experiment_name = os.path.basename(os.path.splitext(filename)[0])
        # Data loading and Exploration
        data, features, targets = load_data(filename, args.target_col)
        create_pairplot(data)
        plt.savefig(path.join(args.output,experiment_name + '_pairplot.png'))
        create_corr_matrix(data)
        plt.savefig(path.join(args.output, experiment_name + '_corr_matrix.png'))

        # Fit model
        reg = train_model(features, targets)
        create_learning_curve(reg, features, targets)
        plt.savefig(path.join(args.output, experiment_name + '_cv_reg_output.png'))

        # Save model
        joblib.dump(reg, path.join(args.output,experiment_name + '_model.sav'))

if __name__ == "__main__":
    main()