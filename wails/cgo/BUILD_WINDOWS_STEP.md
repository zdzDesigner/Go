# cgo应用程序Windows打包步骤

本文档概述了将cgo应用程序打包并分发到Windows系统的步骤。

## 前提条件

在打包之前，请确保您已使用构建脚本构建了应用程序：
```bash
./build.sh
```

## 打包步骤

### 1. 创建发布目录
首先，为Windows发布创建一个目录：
```bash
mkdir -p windows_release
```

### 2. 复制必要文件
将主可执行文件和所有FFmpeg库复制到发布目录：
```bash
cp cgo-app windows_release/
cp -r lib/ffmpeg_output/lib/* windows_release/
```

### 3. 创建辅助脚本
在windows_release目录中创建以下批处理文件：

#### run.bat
```
@echo off
REM 此脚本运行cgo应用程序并设置正确的库路径

REM 获取此批处理文件的目录
set "DIR=%~dp0"

REM 运行应用程序
"%DIR%cgo-app.exe"

REM 暂停以查看任何错误消息
pause
```

#### install.bat
```
@echo off
REM cgo应用程序的Windows安装脚本
REM 此脚本将应用程序文件复制到Program Files目录并创建快捷方式

set "APP_NAME=cgo"
set "INSTALL_DIR=%ProgramFiles%\%APP_NAME%"

echo 正在安装 %APP_NAME%...

REM 创建安装目录
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

REM 将所有文件复制到安装目录
echo 正在复制应用程序文件...
xcopy /E /I /Y "%~dp0*" "%INSTALL_DIR%\"

REM 在开始菜单中创建快捷方式
set "START_MENU=%APPDATA%\Microsoft\Windows\Start Menu\Programs"
if not exist "%START_MENU%\%APP_NAME%" mkdir "%START_MENU%\%APP_NAME%"

echo 正在创建快捷方式...
echo Set oWS = WScript.CreateObject("WScript.Shell") > CreateShortcut.vbs
echo sLinkFile = "%START_MENU%\%APP_NAME%\%APP_NAME%.lnk" >> CreateShortcut.vbs
echo Set oLink = oWS.CreateShortcut(sLinkFile) >> CreateShortcut.vbs
echo oLink.TargetPath = "%INSTALL_DIR%\cgo-app.exe" >> CreateShortcut.vbs
echo oLink.WorkingDirectory = "%INSTALL_DIR%" >> CreateShortcut.vbs
echo oLink.Description = "%APP_NAME% 应用程序" >> CreateShortcut.vbs
echo oLink.Save >> CreateShortcut.vbs
cscript CreateShortcut.vbs
del CreateShortcut.vbs

echo 安装完成！
echo 您现在可以从开始菜单运行 %APP_NAME%。
pause
```

#### uninstall.bat
```
@echo off
REM cgo应用程序的卸载脚本

set "APP_NAME=cgo"
set "INSTALL_DIR=%ProgramFiles%\%APP_NAME%"
set "START_MENU=%APPDATA%\Microsoft\Windows\Start Menu\Programs"

echo 正在卸载 %APP_NAME%...

REM 删除开始菜单快捷方式
if exist "%START_MENU%\%APP_NAME%" (
    rmdir /S /Q "%START_MENU%\%APP_NAME%"
    echo 已删除开始菜单快捷方式
)

REM 删除安装目录
if exist "%INSTALL_DIR%" (
    rmdir /S /Q "%INSTALL_DIR%"
    echo 已删除应用程序文件
)

echo 卸载完成！
pause
```

### 4. 创建文档
创建一个README.txt文件，包含安装和使用说明：

```
# cgo应用程序Windows打包

本文档解释了如何将cgo应用程序打包并分发到Windows系统。

## 包内容

Windows发布包包含：

1. `cgo-app.exe` - 主应用程序可执行文件
2. 所有必需的FFmpeg DLL文件：
   - libavcodec.dll
   - libavdevice.dll
   - libavfilter.dll
   - libavformat.dll
   - libavutil.dll
   - libpostproc.dll
   - libswresample.dll
   - libswscale.dll
3. 辅助脚本：
   - `run.bat` - 运行应用程序
   - `install.bat` - 将应用程序安装到Program Files
   - `uninstall.bat` - 删除应用程序

## 安装选项

### 选项1：直接执行（无需安装）
1. 将包解压到任意目录
2. 双击`run.bat`执行应用程序

### 选项2：完整安装
1. 将包解压到任意目录
2. 右键单击`install.bat`并选择"以管理员身份运行"
3. 应用程序将被安装到`%ProgramFiles%\cgo`
4. 开始菜单中将创建快捷方式

## 卸载

要删除应用程序：
1. 右键单击`uninstall.bat`并选择"以管理员身份运行"
2. 或手动删除`%ProgramFiles%\cgo`目录

## 技术说明

- 应用程序使用动态链接到FFmpeg库以遵守LGPL许可证
- 所有DLL必须与可执行文件在同一目录中或在系统PATH中
- 安装脚本将所有文件复制到中央位置以便于管理
```

### 5. 创建分发存档
最后，创建Windows发布的ZIP存档：
```bash
cd windows_release && zip -r ../cgo-windows.zip *
```

## 使用说明

用户有两种方式运行应用程序：

1. **直接执行**：
   - 解压ZIP文件
   - 双击`run.bat`运行应用程序

2. **完整安装**：
   - 解压ZIP文件
   - 右键单击`install.bat`并选择"以管理员身份运行"
   - 应用程序将被安装到`%ProgramFiles%\cgo`
   - 可以从开始菜单启动

## 技术细节

- 使用动态链接到FFmpeg库以遵守LGPL许可证
- 所有DLL文件必须与可执行文件在同一目录中或在系统PATH中
- 安装脚本将所有文件复制到中央位置以便于管理