import urllib2
import json
import datetime
import time
import pytz
import psycopg2

###
### Define database parameters here
###

host = "pellefant.db.elephantsql.com"
dbname = "ybbhoaeu"
user = "ybbhoaeu"
password = "oscpbdgnEIFVVOrSVieMvqN1RFeJEW6C"

# Set up database connection settings and other parameters
tsOld = str(int(time.time())-180)
hitsPerPage = 1000
conn_string = "host=%s dbname=%s user=%s password=%s" % (host, dbname, user, password)
db = psycopg2.connect(conn_string)
cur = db.cursor()

# Set up HN story database table schema
cur.execute("CREATE TABLE IF NOT EXISTS hn_submissions (objectID int PRIMARY KEY, title varchar, url varchar, author varchar, created_at timestamp);")

num_processed = 0

while True:
	try:
		# Retrieve HN submissions from the Algolia API; finds all submissions before timestamp of last known submission time
		tsNew = str(int(time.time()))
                url = 'http://hn.algolia.com/api/v1/search_by_date?tags=story&hitsPerPage=%s&numericFilters=created_at_i>%s,created_at_i<%s' % (hitsPerPage, tsOld, tsNew)
		req = urllib2.Request(url)
		response = urllib2.urlopen(req)
		
		data = json.loads(response.read())
		submissions = data['hits']
		
                print("")
		for submission in submissions:
		
			# make sure we remove smartquotes/other unicode from title
			title = submission['title'].translate(dict.fromkeys([0x201c, 0x201d, 0x2011, 0x2013, 0x2014, 0x2018, 0x2019, 0x2026, 0x2032])).encode('utf-8')
                        print(title) 
			
			# EST timestamp since USA activity reflects majority of HN activity
			created_at = datetime.datetime.fromtimestamp(int(submission['created_at_i']), tz=pytz.timezone('America/New_York')).strftime('%Y-%m-%d %H:%M:%S')
			
			
			SQL = "INSERT INTO hn_submissions (objectID, title, url, author, created_at) VALUES (%s,%s,%s,%s,%s) ON CONFLICT DO NOTHING"
			insert_data = (int(submission['objectID']), title, submission['url'], submission['author'], created_at)
			
			try:
				cur.execute(SQL, insert_data)
				db.commit()
				
			except Exception, e:
				print insert_data
				print e
		
		# Print the number of processed articles!
                print("number processed: " + str(data["nbHits"]))
		
		# make sure we stay within API limits
		time.sleep(10)
                tsOld = str(int(tsNew)-180)
		
	except Exception, e:
		print e
                time.sleep(10)

# close the db connection
db.close()
