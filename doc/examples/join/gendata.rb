#!/usr/bin/env ruby


$publishers = ["random house", "pelican"]

$authors = ["alice","bob","catie","dan","emma","frank"]
authors_file = File.open("authors.csv", "w")
authors_file << "ID, Name, PublisherID\n"

def rand_selection(set)
    Integer(rand*set.size%set.size)
end

$author_id=0
def new_author(a)
    publisherID = rand_selection($publishers)
    row = [$author_id, a, publisherID]
    $author_id += 1
    row.join(",")
end

$authors.each {|a| authors_file << new_author(a) + "\n"}

# Write books
books_file = File.open("books.csv","w")
books_file << "ID, Title, AuthorID, Price\n"

$book_id=0
def new_book()
    title = (0...50).map { ('a'..'z').to_a[rand(26)] }.join
    price = rand*10
    row = [$book_id, title, rand_selection($authors), sprintf( "%0.02f", price)]
    $book_id += 1
    row.join(",")
end

$book_count = 0
while $book_count < 100
    books_file << new_book() + "\n"
    $book_count += 1
end

# Write sales

def new_sale(txnID)
	[txnID, Integer(rand*$book_count % $book_count)].join(",")
end

batch=0
while batch < 3
    sales_file = File.open("sales#{batch}.csv", "w")
	sales_file << "TransactionID, BookID\n"
    txn=0
    while txn < 100
        sales_file << new_sale(txn) + "\n"
        txn += 1
    end
    batch += 1
end
