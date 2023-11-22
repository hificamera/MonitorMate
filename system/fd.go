package system

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

type Fd struct {
	Used  uint64
	Limit uint64
}

func (this *Fd) Dump() {
	fmt.Printf("Used:%d, Limit:%d", this.Used, this.Limit)
}

func (this *Fd) Collect() error {
	contents, err := ioutil.ReadFile("/proc/sys/fs/file-nr")
	if err != nil {
		return err
	}
	reader := bufio.NewReader(bytes.NewBuffer(contents))

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		fields := strings.Fields(string(line))
		if len(fields) != 3 {
			return errors.New("fd info err")
		}
		this.Used, _ = strconv.ParseUint(fields[0], 10, 64)
		this.Limit, _ = strconv.ParseUint(fields[2], 10, 64)
	}
	return nil
}

func (this *Fd) FdUsed(args string) string {
	return strconv.FormatUint(this.Used, 10)
}

func (this *Fd) FdLimit(args string) string {
	return strconv.FormatUint(this.Limit, 10)
}
