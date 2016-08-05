import cv2
import numpy as np
from matplotlib import pyplot as plt
import os
 
def make_edges(image):
    img = cv2.imread('/pfs/images/' + image)
    edges = cv2.Canny(img,100,200)
    plt.imsave("/pfs/out/" + os.path.splitext(image)[0]+'.png', edges, cmap = 'gray')

for subdir, dirs, files in os.walk("/pfs/images"):
    for file in files:
        make_edges(file)
