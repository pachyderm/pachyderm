#!/usr/local/bin/python3

import glob, os, json, shutil


store_file  = glob.glob(os.path.join("/pfs/stores", "*.txt"))[0]
purchase_file = glob.glob(os.path.join("/pfs/purchases", "*.txt"))[0]

print("Opening store_file...: " + store_file)

with open(store_file, 'r') as store_json:    
    store = json.load(store_json)
    zipcode = store['address']['postal']
    print("zipcode: " + zipcode)
    print("Copying : " + purchase_file + " to: " + "/pfs/out/"+zipcode+"/")
    #os.makedirs("/pfs/out/"+zipcode+"/", exist_ok=True)
    #shutil.copy2(purchase_file, "/pfs/out/"+zipcode+"/") 

    shutil.copy(purchase_file, "/pfs/out/"+zipcode+".txt") 
  

