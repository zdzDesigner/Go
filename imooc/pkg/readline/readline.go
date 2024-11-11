package readline

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

func Read() {
	fmt.Println("main")

	// bts, err := exec.Command("whoami").Output()
	bts, err := exec.Command("udevadm", "info", "-e").Output()
	if err != nil {
		panic(err)
	}
	// read1(bts)
	readLine(bts)
	return
	port, err := serial.Open(serial.OpenOptions{})
	if err != nil {
		fmt.Println(err)
	}
	port.Close()
}

func read1(bts []byte) {
	// fmt.Println(string(bts))
	// fmt.Println(bytes.Fields([]byte("\r")))
	// bytes.NewBuffer(bts).ReadString()
	for {
		time.Sleep(time.Microsecond * 300)
		advance, token, err := bufio.ScanLines(bts, true)
		fmt.Println(advance, string(token), err)
	}
}

func readLine(bts []byte) {
	// reader := bufio.NewReader(strings.NewReader(string(bts)))
	// reader := bufio.NewReader(bytes.NewReader(bts))
	reader := bufio.NewReader(&Serial{bts: bts})

	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		if strings.Contains(string(line), "P:") {
			fmt.Println(string(line))
		}
	}
}

type Serial struct {
	bts []byte
	i   int
}

func (s *Serial) Read(b []byte) (n int, err error) {
	if s.i >= len(s.bts) {
		return 0, io.EOF
	}
	n = copy(b, s.bts[s.i:])
	s.i += n
	return
}
