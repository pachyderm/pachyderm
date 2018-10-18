#!/usr/bin/env ruby

require 'fileutils'

if ARGV.size < 2
    puts "Usage: ./shuffle.rb input_dir output_dir"
    exit 1
end

INPUT_DIR = ARGV[0]
OUTPUT_DIR = ARGV[1]

$author_id_index = 2 # 3rd col contains author id
$pub_id_index = 3 # 4th col contains publisher id

def shuffle_row(type, file, index)
    rows = File.read(file).split("\n")

    FileUtils.mkdir_p File.join(OUTPUT_DIR, type) 
	values = rows.last.split(",")
	output_file_name = File.join(OUTPUT_DIR, type, values[index]+".csv")
	output_file = File.open(output_file_name,"a")
	output_file << rows.last + "\n" # values
end

Dir.glob(File.join(INPUT_DIR, "*")).each do |file|
	puts "shuffling #{file}\n"
	shuffle_row("authors", file, $author_id_index)
	shuffle_row("publishers", file, $pub_id_index)
end
