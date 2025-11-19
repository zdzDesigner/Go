✦ 这个示例程序的功能是演示如何手动创建、配置和操作 `astiav.Frame` 对象，以及如何在 Frame 的内部数据和 Go 的原生数据类型（如 []byte 和
  image.Image）之间进行转换。

  这个示例不涉及任何媒体文件的读写或编解码，它是一个纯粹的 API 功能展示，核心在于 Frame 对象本身。Frame 对象是 FFmpeg
  中用来存放解码后原始数据（一帧视频图像或一段音频采样）的容器。

  程序分为两个主要部分：

  1. 音频帧操作 (Audio Frame Manipulation)

  这部分展示了如何从零开始构建一个音频帧：

   1. 分配对象：使用 astiav.AllocFrame() 创建一个空的 Frame。
   2. 设置属性：手动设置该 Frame 作为音频帧所需的各项属性，包括：
       * 声道布局 (ChannelLayout)，如立体声。
       * 采样数 (NbSamples)。
       * 样本格式 (SampleFormat)，如浮点型。
       * 采样率 (SampleRate)。
   3. 分配内存：调用 AllocBuffer()。FFmpeg 会根据上面设置的属性自动计算并分配足够大的内存来存放音频数据。
   4. 数据交互：
       * 写入：演示了如何将一个 Go 的字节切片 []byte（在实际应用中它会包含原始的 PCM 音频数据）的内容拷贝到 Frame 的内部数据区
         (frame.Data().SetBytes())。
       * 读取：演示了如何将 Frame 内部的数据拷贝出来，转换成一个 Go 的 []byte (frame.Data().Bytes())。

  2. 视频帧操作 (Video Frame Manipulation)

  这部分展示了如何构建一个视频帧，并与 Go 的 image.Image 类型进行交互：

   1. 分配对象：同样，创建一个空的 Frame。
   2. 设置属性：设置作为视频帧所需的属性：
       * 高度 (Height)。
       * 像素格式 (PixelFormat)，如 RGBA。
       * 宽度 (Width)。
   3. 分配内存：调用 AllocBuffer() 来为图像数据分配内存。
   4. 数据交互：
       * 与 `[]byte` 交互：与音频部分类似，演示了如何在 Frame 和 []byte 之间拷贝原始图像数据。
       * 与 `image.Image` 交互（高级功能）：
           * 写入 (`FromImage`)：演示了如何将一个 Go 标准库的 image.Image
             对象（例如从文件中加载的图片或程序生成的图像）的像素数据直接转换并填充到 Frame 中。
           * 读取 (`ToImage`)：演示了如何将 Frame 中的原始像素数据转换成一个 Go 的 image.Image 对象，这样就可以在 Go
             的生态中方便地进行处理，比如保存为 PNG 或 JPEG 文件。

  总结与用途：

  这个示例对于任何需要在 Go 代码中生成媒体内容或深度处理媒体内容的场景都至关重要。例如：

   * 用 Go 代码生成图表或动画，然后将这些图像作为视频帧进行编码。
   * 从视频中解码出一帧，在 Go 中使用图像处理库（如添加滤镜、水印）后，再将处理后的图像放回 Frame 中进行后续编码。
   * 程序化地生成音频信号，并将其编码为音频文件。

