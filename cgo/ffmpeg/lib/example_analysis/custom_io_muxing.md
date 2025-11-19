> @custom_io_muxing/ 这个example 的功能

✦ 这个示例程序的功能与 custom_io_demuxing 相对应，它演示了如何实现 自定义 I/O 复用 (Custom I/O Muxing)。

  简而言之，这个程序从一个媒体文件读取数据，然后将其“转封装”（remux）到另一个容器格式（本例中是 MP4），但在写入输出文件时，它接管了
  FFmpeg 的 I/O 操作，由 Go 程序来执行实际的写入动作。

  具体工作流程如下：

   1. 打开输入文件：程序使用标准方式 (inputFormatContext.OpenInput) 打开一个输入媒体文件，准备从中读取数据包。

   2. 准备自定义输出：
       * 程序使用 Go 的 os.Create() 创建一个输出文件，并获得一个 Go 的文件句柄 (*os.File)。
       * 它初始化了一个用于输出的 OutputFormatContext，指定输出格式为 "mp4"。

   3. 创建可写的 I/O 上下文 (`IOContext`)：
       * 这是核心步骤。程序创建了一个 astiav.IOContext，并将其标记为可写 (writable: true)。
       * 在创建时，它向 IOContext 提供了两个关键的回调函数：
           * 一个写入函数：当 FFmpeg 需要将数据（如文件头、数据包、文件尾）写入输出时，它会调用这个 Go 函数。该函数内部执行的是 Go 的
             file.Write()。
           * 一个寻址函数：某些容器格式在写入过程中可能需要来回移动文件指针，这个函数为此提供了支持。

   4. 关联 I/O 上下文：
       * 创建好的 IOContext 被设置到输出 OutputFormatContext 中。这等于告诉 FFmpeg：“当你需要写入数据时，把数据交给我提供的 Go
         函数处理，不要自己写文件。”

   5. 执行转封装（Remuxing）：
       * 写文件头：调用 outputFormatContext.WriteHeader()。FFmpeg 生成 MP4 文件的头部信息，并通过之前注册的写入回调，由 Go
         程序将其写入文件。
       * 循环读写：程序进入一个循环，不断地：
           1. 从输入文件读取一个数据包（Packet）。
           2. 调整数据包的时间戳（RescaleTs）以匹配输出流。
           3. 调用 outputFormatContext.WriteInterleavedFrame() 将数据包交给 FFmpeg 处理。FFmpeg 会再次通过写入回调，让 Go
              程序把这个数据包的字节内容写入文件。
       * 写文件尾：循环结束后，调用 outputFormatContext.WriteTrailer() 来写入 MP4 文件尾部的重要元数据（如 moov
         atom）。这个过程同样通过自定义的写入回调完成。

  总结与用途：

  这个示例展示了如何控制 FFmpeg 的输出流向。这种能力使得程序可以将编码和复用后的媒体数据发送到任何 Go
  可以写入的目标，而不仅仅是本地文件。实际应用场景包括：


   * 将媒体流直接推送到 RTMP 或 SRT 服务器。
   * 通过 HTTP POST 将数据流式传输到 Web 服务器。
   * 将输出写入内存缓冲区进行进一步处理。
   * 在写入磁盘前对数据进行实时加密。

