# Linux Debian包创建步骤

本文档详细说明了如何为cgo应用程序创建Debian包（.deb文件）。

## 前提条件

在创建包之前，请确保已使用构建脚本构建了应用程序：
```bash
./build.sh
```

## 包结构

Debian包遵循标准结构：
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
│   │   └── (FFmpeg库文件)
│   └── share/
│       ├── applications/
│       │   └── cgo-app.desktop
│       └── doc/
│           └── cgo-app/
│               ├── README
│               └── copyright
```

## 创建包的步骤

### 1. 创建包目录结构
```bash
mkdir -p debian/DEBIAN
mkdir -p debian/usr/bin
mkdir -p debian/usr/lib
mkdir -p debian/usr/share/doc/cgo-app
mkdir -p debian/usr/share/applications
```

### 2. 复制文件
```bash
# 复制主可执行文件
cp cgo-app debian/usr/bin/

# 复制FFmpeg库文件
cp -r lib/ffmpeg_output/lib/* debian/usr/lib/

# 复制文档
cp README.md debian/usr/share/doc/cgo-app/
```

### 3. 创建控制文件
创建`debian/DEBIAN/control`文件，包含包元数据：
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

### 4. 创建桌面入口文件
创建`debian/usr/share/applications/cgo-app.desktop`：
```
[Desktop Entry]
Name=CGO音频处理器
Comment=处理和转换音频文件
Exec=/usr/bin/cgo-app
Icon=cgo-app
Terminal=false
Type=Application
Categories=AudioVideo;Audio;
```

### 5. 创建包脚本
创建`debian/DEBIAN/postinst`：
```bash
#!/bin/bash
set -e

case "$1" in
    configure)
        # 更新桌面数据库
        if [ -x /usr/bin/update-desktop-database ]; then
            update-desktop-database /usr/share/applications
        fi
        
        # 设置正确权限
        chmod 755 /usr/bin/cgo-app
        ;;
esac

exit 0
```

创建`debian/DEBIAN/postrm`：
```bash
#!/bin/bash
set -e

case "$1" in
    purge)
        # 清理配置文件（如果有）
        rm -rf /etc/cgo-app/
        ;;
    remove|upgrade|failed-upgrade|abort-install|abort-upgrade|disappear)
        ;;
esac

exit 0
```

### 6. 设置权限
```bash
chmod 755 debian/DEBIAN/postinst debian/DEBIAN/postrm
```

### 7. 构建包
```bash
dpkg-deb --build debian cgo-app_1.0.0_amd64.deb
```

## 安装包

安装包：
```bash
sudo dpkg -i cgo-app_1.0.0_amd64.deb
sudo apt-get install -f  # 解决依赖问题
```

卸载包：
```bash
sudo dpkg -r cgo-app  # 移除包但保留配置文件
sudo dpkg -P cgo-app  # 完全清除包
```

## 包信息查看

查看包信息：
```bash
dpkg -I cgo-app_1.0.0_amd64.deb  # 显示包信息
dpkg -L cgo-app                   # 列出安装的文件
dpkg -s cgo-app                   # 显示已安装包的状态
```

## 许可证信息

该包包含：
1. MIT许可证下的cgo应用程序
2. LGPLv2.1许可证下的FFmpeg库

版权信息已包含在copyright文件中。

## 故障排除

如果在安装过程中遇到依赖问题：
1. 运行`sudo apt-get install -f`解决依赖关系
2. 如果某些FFmpeg库不可用，可能需要添加额外的软件源或手动安装这些库

## 注意事项

1. 确保所有文件路径正确
2. 控制文件末尾需要有一个空行
3. 脚本文件需要有可执行权限
4. 包版本号应遵循语义化版本控制（major.minor.patch）