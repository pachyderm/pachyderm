#!/usr/bin/env ruby

require 'fileutils'

if ARGV.size < 2
    puts "Usage: ./shuffle.rb input_dir output_dir"
    exit 1
end

INPUT_DIR = ARGV[0]
OUTPUT_DIR = ARGV[1]

def shuffle_row(type, file, index)
    rows = File.read(file).split("\n")

    FileUtils.mkdir_p File.join(OUTPUT_DIR, type) 
	values = rows.last.split(",")
	output_file = File.open(File.join(OUTPUT_DIR, type, values[index] + ".csv"),"w")
	output_file << rows.first # CSV header
	output_file << rows.last # values
end

Dir.glob(File.join(INPUT_DIR, "*")).each do |file|
	author_id_index = 2 # 3rd col contains author id
	pub_id_index = 3 # 4th col contains publisher id

	puts "shuffling #{file}\n"
	shuffle_row(file, "authors", author_id_index)
	shuffle_row(file, "publishers", pub_id_index)
end
