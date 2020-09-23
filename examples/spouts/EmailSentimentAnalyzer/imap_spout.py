import imaplib
import pprint
import os
import tarfile
import errno
import time
import io
import stat


SPOUT = '/pfs/out'

def open_pipe(path_to_file, attempts=0, timeout=2, sleep_int=5):
    if attempts < timeout :
        flags = os.O_WRONLY  # Refer to "man 2 open".
        mode = stat.S_IWUSR  # This is 0o400.
        umask = 0o777 ^ mode  # Prevents always downgrading umask to 0.
        umask_original = os.umask(umask)
        try:
            file = os.open(path_to_file, flags, mode)
            # you must open the pipe as binary to prevent line-buffering problems.
            return os.fdopen(file, "wb")
        except OSError as oe:
            print ('{0} attempt of {1}; error opening file: {2}'.format(attempts + 1, timeout, oe))
            os.umask(umask_original)
            time.sleep(sleep_int)
            return open_pipe(path_to_file, attempts + 1)
        finally:
            os.umask(umask_original)
    return None


unspecified_value = 'not specified';
imap_host = os.getenv('IMAP_SERVER', 'imap.gmail.com')
imap_user = os.getenv('IMAP_LOGIN', unspecified_value)
imap_pass = os.getenv('IMAP_PASSWORD', unspecified_value)
imap_inbox = os.getenv('IMAP_INBOX', 'Inbox')
imap_processed_box = os.getenv('IMAP_PROCESSED_BOX', 'Processed')

if ((imap_pass == unspecified_value) or (imap_user == unspecified_value)):
    print("imap spout error: IMAP_LOGIN and IMAP_PASSWORD environment variables not set.")
    exit(-1)


# connect to host using SSL
imap = imaplib.IMAP4_SSL(imap_host)

## login to server
imap.login(imap_user, imap_pass)

try:
    imap.create(imap_processed_box)
except imaplib.IMAP4.error as im4e:
    print("error creating processed box: {}".format(im4e))
    pass

while (True):
    print("checking for emails...")
    ## select the mailbox for reading messages from
    imap.select(imap_inbox)

    typ, data = imap.uid("search", None, 'ALL')
    all_emails = data[0].split()
    number_of_emails = len(data[0].split())

    if number_of_emails > 0:
        print("{} new emails.".format(number_of_emails))
        mySpout = open_pipe(SPOUT)
        if mySpout is None:
            print ('error opening file: {}'.format(SPOUT))
            exit(-2)

        # To use a tarfile object with a named pipe, you must use the "w|" mode
        # which makes it not seekable
        print("Creating tarstream...")
        try:
            tarStream = tarfile.open(fileobj=mySpout,mode="w|", encoding='utf-8')
        except tarfile.TarError as te:
            print('error creating tarstream: {0}'.format(te))
            exit(-2)

        for current in range(number_of_emails):
            current_uid = all_emails[current]
            typ, email_data = imap.uid("fetch", current_uid, '(RFC822)')
            current_email_rfc822 = email_data[0][1].decode('utf-8')
            name = "{}.mbox".format(current_uid)
            print("Creating tar archive entry for message {}...".format(current_uid))

            tarHeader = tarfile.TarInfo()
            tarHeader.size = len(current_email_rfc822)
            tarHeader.mode = 0o600
            tarHeader.name = name

            print("Writing tarfile to spout for message {}...".format(current_uid))
            try:
                with io.BytesIO(current_email_rfc822.encode('utf-8')) as email:
                    tarStream.addfile(tarinfo=tarHeader, fileobj=email)
            except tarfile.TarError as te:
                print('error writing message {0} to tarstream: {1}'.format(current_uid, te))
                exit(-2)

            print("copying message {} to {}".format(current_uid, imap_processed_box))

            copyResult = imap.uid("copy", current_uid, imap_processed_box)
            if copyResult[0] == "OK":
                print("Deleting message {} from {}".format(current_uid, imap_inbox))
                mov, data = imap.uid("store", current_uid, "+FLAGS", "(\Deleted)")
                imap.expunge()
            else:
                print("Error copying message {} to {}".format(current_uid, imap_processed_box))
                exit(-2)

        tarStream.close()
    else:
        print("No new emails...")

    print("waiting for new emails...")
    time.sleep(5)



mySpout.close()
imap.close()

