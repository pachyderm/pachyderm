#!/usr/bin/env ruby

require 'fileutils'

if ARGV.size < 2
    puts "Usage: ./map.rb input_dir output_dir"
    exit 1
end

INPUT_DIR = ARGV[0]
OUTPUT_DIR = ARGV[1]

Dir.glob(File.join(INPUT_DIR, "*")).each do |table_dir|
    # e.g. sales, books, authors
    table = File.basename(table_dir)
    Dir.glob(File.join(INPUT_DIR, table, "*")).each do |chunk_file|
    # Now we walk over the underlying files added in this commit
        first = true
        File.read(chunk_file).split("\n").each do |line|
            if first #skip header row
                first = false 
                next
            end
            row = line.split(",")
            # Write out e.g. to /pfs/out/authors/0.csv ... the ID of the author
            FileUtils.mkdir_p File.join(OUTPUT_DIR, table) 
            File.open(File.join(OUTPUT_DIR, table, row.first + ".csv"), "w") {|f| f << line}

        end
    end

end
