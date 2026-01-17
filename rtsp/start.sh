#!/bin/bash

# Start the RTSP to WebSocket server
echo "Starting RTSP to WebSocket server..."
./rtsp &

# Wait a moment for server to start
sleep 2

# Open browser
echo "Opening browser..."
if command -v xdg-open &> /dev/null; then
    xdg-open index.html
elif command -v open &> /dev/null; then
    open index.html
elif command -v start &> /dev/null; then
    start index.html
else
    echo "Please open index.html in your browser"
fi

echo "Server running on http://localhost:8080 (WebSocket on ws://localhost:8080/ws)"
echo "Press Ctrl+C to stop"