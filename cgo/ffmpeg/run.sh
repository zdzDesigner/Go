#!/bin/bash

# This script sets the necessary environment for the ffmpeg-concat executable
# and then runs it, passing along all command-line arguments.

# Get the directory where the script is located
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# Set the library path relative to the script's location
export LD_LIBRARY_PATH="$DIR/lib/ffmpeg_output/lib:$LD_LIBRARY_PATH"

# Execute the program with all arguments passed to the script
"$DIR/ffmpeg-concat" "$@"
