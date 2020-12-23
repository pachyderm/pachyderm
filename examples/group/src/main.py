#!/usr/local/bin/python3
import glob, json, os, shutil, sys, csv


store_paths  = glob.glob(os.path.join("/pfs/stores", "*.txt"))
return_paths = glob.glob(os.path.join("/pfs/returns", "*.txt"))
purchase_paths = glob.glob(os.path.join("/pfs/purchases", "*.txt"))  

zipcode = "UNKNOWN"
separator_line = "\nThis return store does not exist \n"
store_id = 0
if store_paths:
    print("Opening store_file...: " + store_paths[0])
    with open(store_paths[0], 'r') as store_json:    
        store = json.load(store_json)
        # Add a text separator to identify in what store the purchase was made
        store_id = store["storeid"]
        separator_line = "\nStore: "+ str(store_id) +" - "+ store["name"]+" \n"   

net_amount = 0.00
order_total = 0.00
return_total = 0.00

for purchase_path in purchase_paths:
    with open(purchase_path, 'r') as purchase:
        csv_reader = csv.reader(purchase, delimiter = '|')
        for row in csv_reader:
            if row[0] == "SKU":
                order_total += float(row[3]) * float(row[5])

for return_path in return_paths:
    with open(return_path, 'r') as return_file:
        csv_reader = csv.reader(return_file, delimiter = '|')
        for row in csv_reader:
            if row[0] == "SKU":
                return_total += float(row[3]) * float(row[5])


with open("/pfs/out/"+str(store_id)+".txt", 'w') as revenue_file:   
    net_amount = order_total - return_total
    revenue_file.write(separator_line)
    revenue_file.write("ORDER_AMOUNT|" + str(order_total) + "|RETURN_AMOUNT|" + str(return_total) + "|NET_AMOUNT|"+ str(net_amount) + "\n")

