from vaderSentiment.vaderSentiment import SentimentIntensityAnalyzer
import os
import email
import mimetypes
from email.policy import default
import argparse

default_input_repo = os.getenv('INPUT_REPO', '/pfs/imap_spout')
default_negatives_dir = os.getenv('NEGATIVES_DIRECTORY', '/pfs/out/negatives')
default_positives_dir = os.getenv('POSITIVES_DIRECTORY', '/pfs/out/positives')
default_sentiment_header = os.getenv('SENTIMENT_HEADER', 'X-Sentiment-Rating')

def gauge_email_sentiment(file, input_dir, positives, negatives, header, analyzer):
    with open(os.path.join(input_dir,file), 'rb') as email_fp:
        msg = email.message_from_binary_file(email_fp, policy=default)
    msg_body = msg.get_body(preferencelist=('plain'))
    # We include the subject in scoring the message
    score = analyzer.polarity_scores("{} {}".format(msg['subject'], msg_body.get_content()))
    # This would score without the subject
    # score = analyzer.polarity_scores(msg_body.get_content())
    # Put the scores in the envelope for later use
    msg.add_header(header, str(score))
    # Decide where to put the message
    if score['compound'] < 0:
        output_path = negatives
    else:
        output_path = positives
    # Write the message out, using the same filename
    with open(os.path.join(output_path,file), "wb") as out_fp:
        out_fp.write(msg.as_bytes())

def main():
        parser = argparse.ArgumentParser(description='Unpack each of the email messages found in a directory (default /pfs/imap_spout), grab the subject and plain text, rate it for sentiments, add a header (default X-Sentiment-Rating) with the rating, and sort into one of two directories based on positive (default /pfs/out/positive) or negative sentiment (default /pfs/out/negative) ratings.')

        parser.add_argument('-i', '--input_repo', required=False,
                            help="""The directory where the emails to be processed are to be found, one email per file. This overrides the default and the environment variable INPUT_REPO.""",
                            default=default_input_repo)
        parser.add_argument('-n', '--negatives_dir', required=False,
                            help="""Where the negative emails go. This overrides the default and the environment variable NEGATIVES_DIRECTORY.""", default=default_negatives_dir)
        parser.add_argument('-p', '--positives_dir', required=False,
                            help="""Where the positive emails go. This overrides the default and the environment variable POSITIVES_DIRECTORY.""", default=default_positives_dir)
        parser.add_argument('-s', '--sentiment_header', required=False,
                            help="""The header that gets the full sentiment rating on the output email. This overrides the default and the environment variable SENTIMENT_HEADER.""", default=default_sentiment_header)
        args = parser.parse_args()
        analyzer = SentimentIntensityAnalyzer()

        try:
            os.mkdir(args.negatives_dir)
        except FileExistsError:
            pass
    
        try:
            os.mkdir(args.positives_dir)
        except FileExistsError:
            pass

        for dirpath, dirs, files in os.walk(args.input_repo):
            for file in files:
                gauge_email_sentiment(file, dirpath, args.positives_dir, args.negatives_dir, args.sentiment_header, analyzer)


if __name__== "__main__":
  main()
