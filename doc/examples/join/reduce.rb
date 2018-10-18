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
	output_file = File.open(File.join(OUTPUT_DIR, table_dir), "w")

	# Read all rows
	rows = []
	header = nil
    Dir.glob(File.join(table_dir, "*")).each do |file|
		# No need to read into mem just to append, just use bash
		raw = File.read(file).split("\n")
		header = raw.first
		rows << raw.last.split(",") # Don't want the header
    end

	# Sort the result by price
	rows.sort_by! {|row| row[$price_index]}

	# Output the sorted result
	output_file << header
	rows.each do |row|
		output_file << row.join(",") + "\n"
	end
end
