import imaplib
import pprint
import os
import tarfile
import errno
import time

SPOUT = '/pfs/out'

def open_pipe(path_to_file, attempts=0, timeout=2, sleep_int=5):
    if attempts < timeout : 
        try:
            file = os.open(path_to_file, os.O_WRONLY)
            return file
        except OSError as oe:
            print ('{0} attempt of {1}; error opening file: {2}'.format(attempts + 1, timeout, oe))
            time.sleep(sleep_int)
            return open_pipe(path_to_file, attempts + 1)
    else:
        return None


unspecified_value = 'not specified';
imap_host = os.getenv('IMAP_SERVER', 'imap.gmail.com')
imap_user = os.getenv('IMAP_LOGIN', unspecified_value)
imap_pass = os.getenv('IMAP_PASSWORD', unspecified_value)
imap_mailbox = os.getenv('IMAP_MAILBOX', 'Inbox')

if ((imap_pass == unspecified_value) or (imap_user == unspecified_value)):
    print("imap spout error: IMAP_LOGIN and IMAP_PASSWORD environment variables not set.")
    exit(-1)


# connect to host using SSL
imap = imaplib.IMAP4_SSL(imap_host)

## login to server
imap.login(imap_user, imap_pass)

## select the mailbox for reading messages from
imap.select(imap_mailbox)


typ, data = imap.search(None, 'ALL')
for num in data[0].split():
    typ, data = imap.fetch(num, '(RFC822)')
    mySpout = open_pipe(SPOUT)
    if mySpout is None:
        print ('error opening file: {}'.format(SPOUT))
        exit(-2)
    else:
        print("Writing to spout for message {}...".format(num))
        os.write(mySpout, data[0][1])

#    print('Message {0}\n{1}\n'.format(num, data[0][1]))
#    os.close(mySpout);
            

# tmp, data = imap.search(None, 'NEW')

# for num in data[0].split():
#     tmp, data = imap.fetch(num, '(RFC822)')
#     print('Message: {0}\n'.format(num))
#     pprint.pprint(data[0][1])
#     break

imap.close()

