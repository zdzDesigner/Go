 2. 使用 rpath (Run-time Search Path) - 推荐

  这是解决你问题的最标准、最简单的方法。

   * 原理:
     在编译可执行文件时，在文件内部嵌入一个特殊的“路径”（rpath）。当程序启动时，操作系统会首先在这个嵌入的路径里查找所需的
     .so 动态库文件。
   * 优点:
       * 你的可执行文件仍然是动态链接的，体积较小。
       * 它“知道”去哪里找自己的依赖库，无需 LD_LIBRARY_PATH。
       * 我们可以使用一个特殊变量
         $ORIGIN，它代表“可执行文件所在的目录”。这样，无论你把项目移动到哪里，它都能正确找到相对路径下的库文件。

  如何实现 rpath

  我们需要在 go build 命令中添加一个链接器标志（ldflag）来设置 rpath。对于你的项目结构，这个命令看起来像这样：











## LD_DEBUG

➜  ffmpeg git:(feature/build) ✗ LD_DEBUG=libs ./ffmpeg-concat 2>&1 | grep "libavcodec" 
    365588:     find library=libavcodec.so.61 [0]; searching 
    365588:       trying file=tls/haswell/x86_64/libavcodec.so.61 
    365588:       trying file=tls/haswell/libavcodec.so.61 
    365588:       trying file=tls/x86_64/libavcodec.so.61 
    365588:       trying file=tls/libavcodec.so.61 
    365588:       trying file=haswell/x86_64/libavcodec.so.61 
    365588:       trying file=haswell/libavcodec.so.61 
    365588:       trying file=x86_64/libavcodec.so.61 
    365588:       trying file=libavcodec.so.61 
    365588:       trying file=/home/zdz/Documents/Speech/meeting/aimt-heming/packages/ffi/TestFFI/tls/haswell/x86_64/libavcodec.so.61 
    365588:       trying file=/home/zdz/Documents/Speech/meeting/aimt-heming/packages/ffi/TestFFI/tls/haswell/libavcodec.so.61 
    365588:       trying file=/home/zdz/Documents/Speech/meeting/aimt-heming/packages/ffi/TestFFI/tls/x86_64/libavcodec.so.61 
    365588:       trying file=/home/zdz/Documents/Speech/meeting/aimt-heming/packages/ffi/TestFFI/tls/libavcodec.so.61 
    365588:       trying file=/home/zdz/Documents/Speech/meeting/aimt-heming/packages/ffi/TestFFI/haswell/x86_64/libavcodec.so.61 
    365588:       trying file=/home/zdz/Documents/Speech/meeting/aimt-heming/packages/ffi/TestFFI/haswell/libavcodec.so.61 
    365588:       trying file=/home/zdz/Documents/Speech/meeting/aimt-heming/packages/ffi/TestFFI/x86_64/libavcodec.so.61 
    365588:       trying file=/home/zdz/Documents/Speech/meeting/aimt-heming/packages/ffi/TestFFI/libavcodec.so.61 
    365588:       trying file='/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib'/tls/haswell/x86_64/libavcodec.so.61 
    365588:       trying file='/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib'/tls/haswell/libavcodec.so.61 
    365588:       trying file='/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib'/tls/x86_64/libavcodec.so.61 
    365588:       trying file='/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib'/tls/libavcodec.so.61 
    365588:       trying file='/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib'/haswell/x86_64/libavcodec.so.61 
    365588:       trying file='/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib'/haswell/libavcodec.so.61 
    365588:       trying file='/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib'/x86_64/libavcodec.so.61 
    365588:       trying file='/home/zdz/Documents/Try/Go/cgo/ffmpeg/lib/ffmpeg_output/lib'/libavcodec.so.61 
    365588:       trying file=/lib/x86_64-linux-gnu/tls/haswell/x86_64/libavcodec.so.61 
    365588:       trying file=/lib/x86_64-linux-gnu/tls/haswell/libavcodec.so.61 
    365588:       trying file=/lib/x86_64-linux-gnu/tls/x86_64/libavcodec.so.61 
    365588:       trying file=/lib/x86_64-linux-gnu/tls/libavcodec.so.61 
    365588:       trying file=/lib/x86_64-linux-gnu/haswell/x86_64/libavcodec.so.61 
    365588:       trying file=/lib/x86_64-linux-gnu/haswell/libavcodec.so.61 
    365588:       trying file=/lib/x86_64-linux-gnu/x86_64/libavcodec.so.61 
    365588:       trying file=/lib/x86_64-linux-gnu/libavcodec.so.61 
    365588:       trying file=/usr/lib/x86_64-linux-gnu/tls/haswell/x86_64/libavcodec.so.61 
    365588:       trying file=/usr/lib/x86_64-linux-gnu/tls/haswell/libavcodec.so.61 
    365588:       trying file=/usr/lib/x86_64-linux-gnu/tls/x86_64/libavcodec.so.61 
    365588:       trying file=/usr/lib/x86_64-linux-gnu/tls/libavcodec.so.61 
    365588:       trying file=/usr/lib/x86_64-linux-gnu/haswell/x86_64/libavcodec.so.61 
    365588:       trying file=/usr/lib/x86_64-linux-gnu/haswell/libavcodec.so.61 
    365588:       trying file=/usr/lib/x86_64-linux-gnu/x86_64/libavcodec.so.61 
    365588:       trying file=/usr/lib/x86_64-linux-gnu/libavcodec.so.61 
    365588:       trying file=/lib/tls/haswell/x86_64/libavcodec.so.61 
    365588:       trying file=/lib/tls/haswell/libavcodec.so.61 
    365588:       trying file=/lib/tls/x86_64/libavcodec.so.61 
    365588:       trying file=/lib/tls/libavcodec.so.61 
    365588:       trying file=/lib/haswell/x86_64/libavcodec.so.61 
    365588:       trying file=/lib/haswell/libavcodec.so.61 
    365588:       trying file=/lib/x86_64/libavcodec.so.61 
    365588:       trying file=/lib/libavcodec.so.61 
    365588:       trying file=/usr/lib/tls/haswell/x86_64/libavcodec.so.61 
    365588:       trying file=/usr/lib/tls/haswell/libavcodec.so.61 
    365588:       trying file=/usr/lib/tls/x86_64/libavcodec.so.61 
    365588:       trying file=/usr/lib/tls/libavcodec.so.61 
    365588:       trying file=/usr/lib/haswell/x86_64/libavcodec.so.61 
    365588:       trying file=/usr/lib/haswell/libavcodec.so.61 
    365588:       trying file=/usr/lib/x86_64/libavcodec.so.61 
    365588:       trying file=/usr/lib/libavcodec.so.61 
./ffmpeg-concat: error while loading shared libraries: libavcodec.so.61: cannot open shared object file: No such file or directory 

