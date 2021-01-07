#!/usr/local/bin/python3
import glob, json, os, shutil, sys


store_paths  = glob.glob(os.path.join("/pfs/stores", "*.txt"))
return_paths = glob.glob(os.path.join("/pfs/returns", "*.txt"))
  

zipcode = "UNKNOWN"
separator_line = "\nThis return store does not exist \n"
if store_paths:
    print("Opening store_file...: " + store_paths[0])
    with open(store_paths[0], 'r') as store_json:    
        store = json.load(store_json)
        zipcode = store['address']['zipcode']
        # Add a text separator to identify in what store the purchase was made
        separator_line = "\nReturn at store: "+ str(store["storeid"]) +" - "+ store["name"]+" \n"   
print("zipcode: " + zipcode)
    
with open("/pfs/out/"+zipcode+".txt", 'w') as location_file:   
    location_file.write(separator_line)
    if return_paths:  
        print("Appending : " + return_paths[0] + " to:" + "/pfs/out/"+zipcode+".txt")
        # Copy the content of the purchase file into the corresponding zipcode file
        with  open(return_paths[0], 'r') as return_file:
            location_file.write(return_file.read()) 




