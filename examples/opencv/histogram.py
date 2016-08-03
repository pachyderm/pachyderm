import cv2
import numpy as np
from matplotlib import pyplot as plt
import os
 
def make_histogram(image):
     img = cv2.imread('/pfs/images/' + image)
     color = ('b','g','r')
     for i,col in enumerate(color):
         histr = cv2.calcHist([img],[i],None,[256],[0,256])
         plt.plot(histr,color = col)
         plt.xlim([0,256])
     plt.savefig("/pfs/out/" + os.path.splitext(image)[0]+'.png')
     plt.clf()

for subdir, dirs, files in os.walk("/pfs/images"):
    for file in files:
        make_histogram(file)
