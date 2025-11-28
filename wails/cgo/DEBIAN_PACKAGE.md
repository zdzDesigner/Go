# Debian Package Creation for cgo Application

This document explains how to create a Debian package (.deb) for the cgo application.

## Prerequisites

Before creating the package, ensure you have built the application using the build script:
```bash
./build.sh
```

## Package Structure

The Debian package follows the standard structure:
```
debian/
├── DEBIAN/
│   ├── control
│   ├── postinst
│   └── postrm
├── usr/
│   ├── bin/
│   │   └── cgo-app
│   ├── lib/
│   │   └── (FFmpeg libraries)
│   └── share/
│       ├── applications/
│       │   └── cgo-app.desktop
│       └── doc/
│           └── cgo-app/
│               ├── README
│               └── copyright
```

## Creating the Package

### 1. Create Package Structure
```bash
mkdir -p debian/DEBIAN
mkdir -p debian/usr/bin
mkdir -p debian/usr/lib
mkdir -p debian/usr/share/doc/cgo-app
mkdir -p debian/usr/share/applications
```

### 2. Copy Files
```bash
# Copy main executable
cp cgo-app debian/usr/bin/

# Copy FFmpeg libraries
cp -r lib/ffmpeg_output/lib/* debian/usr/lib/

# Copy documentation
cp README.md debian/usr/share/doc/cgo-app/
```

### 3. Create Control File
Create `debian/DEBIAN/control` with package metadata:
```
Package: cgo-app
Version: 1.0.0
Section: sound
Priority: optional
Architecture: amd64
Depends: libavcodec58, libavformat58, libavutil56, libswscale5, libavdevice58, libavfilter7, libswresample3
Maintainer: Your Name <your.email@example.com>
Homepage: https://github.com/yourusername/cgo-app
Description: Audio processing application using FFmpeg
 A GUI application for audio processing that uses FFmpeg libraries for
 audio conversion and manipulation. Built with Go and Wails framework.
 .
 Features:
  - Concatenate multiple audio files
  - Convert between various audio formats
  - Progress tracking during processing
```

### 4. Create Desktop Entry
Create `debian/usr/share/applications/cgo-app.desktop`:
```
[Desktop Entry]
Name=CGO Audio Processor
Comment=Process and convert audio files
Exec=/usr/bin/cgo-app
Icon=cgo-app
Terminal=false
Type=Application
Categories=AudioVideo;Audio;
```

### 5. Create Package Scripts
Create `debian/DEBIAN/postinst`:
```bash
#!/bin/bash
set -e

case "$1" in
    configure)
        # Update desktop database
        if [ -x /usr/bin/update-desktop-database ]; then
            update-desktop-database /usr/share/applications
        fi
        
        # Set proper permissions
        chmod 755 /usr/bin/cgo-app
        ;;
esac

exit 0
```

Create `debian/DEBIAN/postrm`:
```bash
#!/bin/bash
set -e

case "$1" in
    purge)
        # Clean up configuration files if any
        rm -rf /etc/cgo-app/
        ;;
    remove|upgrade|failed-upgrade|abort-install|abort-upgrade|disappear)
        ;;
esac

exit 0
```

### 6. Set Permissions
```bash
chmod 755 debian/DEBIAN/postinst debian/DEBIAN/postrm
```

### 7. Build Package
```bash
dpkg-deb --build debian cgo-app_1.0.0_amd64.deb
```

## Installing the Package

To install the package:
```bash
sudo dpkg -i cgo-app_1.0.0_amd64.deb
sudo apt-get install -f  # To resolve any dependency issues
```

To remove the package:
```bash
sudo dpkg -r cgo-app  # Remove package but keep config files
sudo dpkg -P cgo-app  # Purge package completely
```

## Package Information

To check package information:
```bash
dpkg -I cgo-app_1.0.0_amd64.deb  # Show package info
dpkg -L cgo-app                   # List installed files
dpkg -s cgo-app                   # Show installed package status
```

## Licensing

This package includes:
1. The cgo application under the MIT license
2. FFmpeg libraries under LGPLv2.1

Proper attribution is included in the copyright file.