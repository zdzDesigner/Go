#!/bin/bash


export ANDROID_API=21
export ANDROID_NDK_HOME=/home/zdz/Android/Sdk/ndk/21.4.7075529
export PATH=${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/linux-x86_64/bin:${PATH}
export ANDROID_SYSROOT=${ANDROID_NDK_HOME}/toolchains/llvm/prebuilt/linux-x86_64/sysroot
export ANDROID_TOOLCHAIN=${ANDROID_NDK_HOME}/toolchains/arm-linux-androideabi-4.9/prebuilt/linux-x86_64
# export JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
# export ANDROID_USER_HOME=/home/zdz/Android/Sdk


## 系统环境变量清除
# export ANDROID_PREFS_ROOT=/home/zdz/Android/Sdk





CC="armv7a-linux-androideabi${ANDROID_API}-clang" \
CGO_CFLAGS="-I${ANDROID_SYSROOT}/usr/include -I${ANDROID_SYSROOT}/usr/include/arm-linux-androideabi --sysroot=${ANDROID_SYSROOT} -D__ANDROID_API__=${ANDROID_API}" \
CGO_LDFLAGS="-L${ANDROID_SYSROOT}/usr/lib/arm-linux-androideabi/${ANDROID_API} -L${ANDROID_TOOLCHAIN}/arm-linux-androideabi/lib --sysroot=${ANDROID_SYSROOT}" \
CGO_ENABLED=1 GOOS=android GOARCH=arm \
go build -buildmode=c-shared -ldflags="-s -w -extldflags=-Wl,-soname,libexample.so" -o=android/libs/armeabi-v7a/libexample.so


java --version

echo 'build...'
bash ./gradlew assembleDebug --stacktrace


adb uninstall "com.example.goraylib"
adb install ./android/build/outputs/apk/debug/android-debug.apk
