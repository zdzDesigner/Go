package service

import (
	"fmt"
	"testing"
)

func Test_ParseSDP(t *testing.T) {
	sdp_str := `
SDP:
v=0
o=- 0 0 IN IP4 127.0.0.1
s=No Name
c=IN IP4 127.0.0.1
t=0 0
a=tool:libavformat 58.29.100
m=video 5004 RTP/AVP 96
a=rtpmap:96 H264/90000
a=fmtp:96 packetization-mode=1
`
	sdp_info, err := ParseSDP(sdp_str)
	fmt.Println(sdp_info, err)
}
