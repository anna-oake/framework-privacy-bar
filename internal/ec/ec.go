package ec

import (
	"fmt"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const (
	ecCmdPrivacySwitches = 0x3e14
	crosECDevIOCXCmd     = 0xc014ec00 // _IOWR(0xEC, 0, struct cros_ec_command), sizeof header == 20
	ecResultSuccess      = 0
)

type crosECCommand struct {
	Version uint32
	Command uint32
	Outsize uint32
	Insize  uint32
	Result  uint32
	Data    [2]byte
}

func PollMicrophonePrivacy(pollInterval time.Duration) (<-chan bool, error) {
	f, err := os.OpenFile("/dev/cros_ec", os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	ch := make(chan bool)
	go func() {
		defer f.Close()

		last := false
		haveLast := false

		for {
			microphoneConnected, err := readMicrophonePrivacy(f.Fd())
			if err != nil {
				fmt.Println("poll error:", err)
			} else if !haveLast || microphoneConnected != last {
				ch <- microphoneConnected
				last = microphoneConnected
				haveLast = true
			}
			time.Sleep(pollInterval)
		}
	}()

	return ch, nil
}

func readMicrophonePrivacy(fd uintptr) (bool, error) {
	cmd := crosECCommand{
		Command: ecCmdPrivacySwitches,
		Insize:  2,
	}

	n, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		crosECDevIOCXCmd,
		uintptr(unsafe.Pointer(&cmd)),
	)
	if errno != 0 {
		return false, errno
	}
	if cmd.Result != ecResultSuccess {
		return false, fmt.Errorf("EC result %d", cmd.Result)
	}
	if n != 2 {
		return false, fmt.Errorf("short EC response: %d bytes", n)
	}
	if cmd.Data[0] > 1 {
		return false, fmt.Errorf("invalid EC microphone privacy response: %d", cmd.Data[0])
	}

	return cmd.Data[0] == 1, nil
}
