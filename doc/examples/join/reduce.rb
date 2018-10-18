#!/usr/bin/env ruby

require 'fileutils'

if ARGV.size < 2
    puts "Usage: ./reduce.rb input_dir output_dir"
    exit 1
end

INPUT_DIR = ARGV[0]
OUTPUT_DIR = ARGV[1]

$price_index = 4

Dir.glob(File.join(INPUT_DIR, "*")).each do |table_dir|
    # Either data indexed by publisher or author
    table = File.basename(table_dir)
	output_file = File.open(File.join(OUTPUT_DIR, table) + ".csv", "w")

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
end
