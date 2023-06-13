#!/bin/bash

# Set the destination directory where you want to copy the .proto files
DEST_DIR="all-protos"

# Create the destination directory if it doesn't exist
mkdir -p "$DEST_DIR"

# Find all .proto files recursively in the source directory
proto_files=($(find . -name "*.proto"))

# Iterate over each .proto file
for file in "${proto_files[@]}"; do
    # Get the file name and directory
    filename=$(basename "$file")

    # Copy the file to the destination directory with the original name and directory structure
    cp "$file" "$DEST_DIR/$filename"
done

echo "Proto files copied to $DEST_DIR"
