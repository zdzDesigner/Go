#!/bin/bash

# scp zdz@192.168.0.105:/home/zdz/Documents/Try/Zig/zig-pro/windt/**/*.zig ./
# rsync -avz zdz@192.168.0.105:/home/zdz/Documents/Try/Zig/zig-pro/windt/{build.zig,build.zig.zon,pull.sh} ./
# rsync -avz --exclude='zig-out/' zdz@192.168.0.104:/home/zdz/Documents/Try/Zig/zig-pro/windt/* ./
# rsync -avz --exclude='.zig-cache/' --exclude='zig-out/' zdz@192.168.0.105:/home/zdz/Documents/Try/Zig/zig-pro/windt/* ./
 rsync -avz --exclude='.zig-cache/' --exclude='zig-out/' zdz@172.16.57.32:/home/zdz/Documents/Try/Go/mqtt/custom_client/* ./


