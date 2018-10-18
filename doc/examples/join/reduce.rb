#!/usr/bin/env ruby

require 'fileutils'

if ARGV.size < 2
    puts "Usage: ./reduce.rb input_dir output_dir"
    exit 1
end

INPUT_DIR = ARGV[0]
OUTPUT_DIR = ARGV[1]

Dir.glob(File.join(INPUT_DIR, "*")).each do |table_dir|
    # Either data indexed by publisher or author
    table = File.basename(table_dir)
    FileUtils.mkdir_p File.join(OUTPUT_DIR, "rankings") 
	output_file = File.open(File.join(OUTPUT_DIR, "rankings", table) + ".csv", "w")

	# Read all rows
	rows = []
	sum = 0.00
    Dir.glob(File.join(table_dir, "*")).each do |file|
		id = File.basename(file).split(".").first # Filename is object ID (author / publisher)
		sales = File.read(file).split("\n")
		sales.each do |sale|
			sum += sale.split(",").last.to_f	
		end
		rows << [id, sum]
    end

	# Sort the result by price
	rows.sort_by! {|row| row.last}

	# Output the sorted result
	rows.reverse.each do |row|
		output_file << row.join(",") + "\n"
	end

	top_seller = rows.reverse.first

	# Cross reference the top result w the original table to get a name
    FileUtils.mkdir_p File.join(OUTPUT_DIR, "top_seller") 
	output_file = File.open(File.join(OUTPUT_DIR, "top_seller", table) + ".txt", "w")
	File.read(File.join("/pfs/data", table, "1.csv")).split("\n").each do |line|
		row = line.split(",")
		if row.first == top_seller.first
			output_file << row[1] + "\n" # Both schemas have name as the 2nd field
		end
	end
end
