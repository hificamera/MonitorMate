package main

import (
	"fmt"
	"system-info/system"
	"time"
)

func refresh(length int) {
	fmt.Print('\r')

}

func main() {
	mem := &system.Mem{}
	cpu := &system.Cpu{}
	diskIO := &system.DiskIO{}
	disk := &system.Disk{}
	fd := &system.Fd{}
	// netWork := &system.NetWork{}

	timer := time.NewTicker(1 * time.Second)
	title := "cpu_used\tcpu_idle\tmem_used\tmem_used_rate\tio_read\tio_write\tdisk_used_rate\tfd_used\tfd_limit"
	fmt.Println(title)
	length := len(title)
	refreshStr := "\r"
	for i := 0; i < length; i++ {
		refreshStr = refreshStr + " "
	}
	for range timer.C {
		cpu.Collect()
		mem.Collect()
		diskIO.Collect()
		disk.Collect()
		fd.Collect()
		// netWork.Collect()
		fmt.Printf(refreshStr)
		fmt.Printf("\r%s%%\t\t%s%%\t\t%sKB\t\t%s%%\t%sKB\t%sKB\t\t%s%%\t\t%s\t%s",
			cpu.UserRateFunc(""),
			cpu.IdleRateFunc(""),
			mem.MemUsedFunc(""),
			mem.MemUsedRateFunc(""),
			diskIO.RkbPerSecondFunc(""),
			diskIO.WkbPerSecondFunc(""),
			disk.DiskUsedRate(""),
			fd.FdUsed(""),
			fd.FdLimit(""),
		)
	}
}
