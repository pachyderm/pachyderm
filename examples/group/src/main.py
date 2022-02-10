#!/usr/local/bin/python3
import glob, json, os, shutil, sys, csv

store_id = os.environ.get('PACH_DATUM_stores_GROUP_BY')
store_id = str(0) if store_id == None else str(store_id) 


store_path  = os.path.join("/pfs/stores","STOREID"+store_id+".txt")
return_paths = glob.glob(os.path.join("/pfs/returns", "*.txt"))
purchase_paths = glob.glob(os.path.join("/pfs/purchases", "*.txt"))  

zipcode = "UNKNOWN"
separator_line = "\nThis return store does not exist \n"

if store_id != "0":
    print("Opening store_file...: " + store_id)
    with open(store_path, 'r') as store_json:  
        store = json.load(store_json)
        # Add a text separator to identify in what store the purchase was made
        separator_line = "\nStore: "+ store_id +" - "+ store["name"]+" \n"   

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


## Create directory with zipcode name
with open("/pfs/out/"+store_id+".txt", 'w') as revenue_file:   
    net_amount = order_total - return_total
    revenue_file.write(separator_line)
    revenue_file.write("\nORDER_AMOUNT|" + str(order_total) + "|RETURN_AMOUNT|" + str(return_total) + "|NET_AMOUNT|"+ str(net_amount) + "\n")

