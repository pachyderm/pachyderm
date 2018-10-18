#!/usr/bin/env ruby

if ARGV.size < 2
    puts "Usage: ./shuffle.rb input_dir output_dir"
    exit 1
end

INPUT_DIR = ARGV[0]
OUTPUT_DIR = ARGV[1]


def read_row(type)
	filename = Dir.glob(File.join(INPUT_DIR, type, type, "*")).first
	File.read(filename).split(",")
end


def join(sale, book, author)
	# Want: SaleID, BookID, AuthorID, PubID, Price
	cols = ["saleID", "bookID", "authorID", "publisherID", "bookPrice"]
	data = [sale[0], book[0], author[0], author[2], book[3]]
	return cols, data
end

sale = read_row("sales")
author = read_row("authors")
book = read_row("books")

if sale[1] == book[0] && book[2] == author[0]
	f = File.open(File.join(OUTPUT_DIR, sale[0] + ".csv"), "w") 
	cols, data = join(sale, book, author)
	f << cols.join(",") + "\n"
	f << data.join(",")
end
