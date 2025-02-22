package service

import (
	"strconv"
	"strings"
)

type SDPInfo struct {
	Version     string
	Origin      string
	SessionName string
	Connection  string
	Timing      string
	Tool        string
	Media       []MediaInfo
}

type MediaInfo struct {
	Type             string
	Port             int
	Protocol         string
	PayloadType      int
	RTPMap           map[int]string
	FormatParameters map[int]string
}

func ParseSDP(data []byte) (*SDPInfo, error) {
	// data, err := os.ReadFile(filePath)
	// if err != nil {
	// 	return nil, fmt.Errorf("读取SDP文件失败: %v", err)
	// }
	//
	lines := strings.Split(string(data), "\n")
	sdpInfo := &SDPInfo{
		Media: make([]MediaInfo, 0),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		switch line[0] {
		case 'v':
			sdpInfo.Version = line[2:]
		case 'o':
			sdpInfo.Origin = line[2:]
		case 's':
			sdpInfo.SessionName = line[2:]
		case 'c':
			sdpInfo.Connection = line[2:]
		case 't':
			sdpInfo.Timing = line[2:]
		case 'a':
			parts := strings.Split(line[2:], ":")
			if len(parts) < 2 {
				continue
			}
			keyVal := strings.Split(parts[1], " ")
			if len(keyVal) < 2 {
				continue
			}
			attrName := keyVal[0]
			attrValue := strings.Join(keyVal[1:], " ")
			if strings.HasPrefix(attrName, "rtpmap") {
				// 处理rtpmap
				mediaType := strings.Split(attrValue, " ")[0]
				payloadType := strings.Trim(strings.Split(attrValue, " ")[0], "=")
				pt, _ := strconv.Atoi(payloadType)
				sdpInfo.Media[len(sdpInfo.Media)-1].RTPMap[pt] = mediaType
			} else if strings.HasPrefix(attrName, "fmtp") {
				// 处理fmtp
				params := strings.Split(attrValue, " ")
				for _, param := range params {
					kv := strings.Split(param, "=")
					if len(kv) != 2 {
						continue
					}
					key := kv[0]
					value := kv[1]
					sdpInfo.Media[len(sdpInfo.Media)-1].FormatParameters[len(sdpInfo.Media[len(sdpInfo.Media)-1].FormatParameters)] = key + "=" + value
				}
			}
		case 'm':
			parts := strings.Split(line[2:], " ")
			if len(parts) < 4 {
				continue
			}
			mediaType := parts[0]
			port, _ := strconv.Atoi(parts[1])
			protocol := parts[2]
			payloadType, _ := strconv.Atoi(parts[3])
			mediaInfo := MediaInfo{
				Type:             mediaType,
				Port:             port,
				Protocol:         protocol,
				PayloadType:      payloadType,
				RTPMap:           make(map[int]string),
				FormatParameters: make(map[int]string),
			}
			sdpInfo.Media = append(sdpInfo.Media, mediaInfo)
		}
	}

	return sdpInfo, nil
}
