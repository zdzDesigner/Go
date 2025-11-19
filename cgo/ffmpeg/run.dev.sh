export CGO_CFLAGS="-I{{ path to your ffmpeg directory }}/include/"
export CGO_LDFLAGS="-L{{ path to your ffmpeg directory }}/lib/"
export PKG_CONFIG_PATH="{{ path to your ffmpeg directory }}/lib/pkgconfig"
