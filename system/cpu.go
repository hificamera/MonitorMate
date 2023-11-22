package system

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

//CPU型号
func CpuModel(args string) string {
	return ExecOutput("cat /proc/cpuinfo | grep name | cut -f2 -d: | uniq")
}

//逻辑CPU个数
func CpuNum(args string) string {
	return ExecOutput("cat /proc/cpuinfo | grep processor | wc -l")
}

type CpuTime struct {
	User    uint64 //从系统启动开始累计到当前时刻，用户态的CPU时间（单位：jiffies） ，不包含 nice值为负进程。1jiffies=0.01秒
	Nice    uint64 //从系统启动开始累计到当前时刻，nice值为负的进程所占用的CPU时间（单位：jiffies）
	System  uint64 //从系统启动开始累计到当前时刻，内核态时间（单位：jiffies）
	Idle    uint64 //从系统启动开始累计到当前时刻，除硬盘IO等待时间以外其它等待时间（单位：jiffies)
	Iowait  uint64 //从系统启动开始累计到当前时刻，硬盘IO等待时间（单位：jiffies）
	Irq     uint64 //从系统启动开始累计到当前时刻，硬中断时间（单位：jiffies）
	SoftIrq uint64 //从系统启动开始累计到当前时刻，软中断时间（单位：jiffies）
	Steal   uint64
	Total   uint64 //user + nice + system + idle + iowait
}

type Cpu struct {
	CpuTimes            []CpuTime
	User                uint64 //从系统启动开始累计到当前时刻，用户态的CPU时间（单位：jiffies） ，不包含 nice值为负进程。1jiffies=0.01秒
	Nice                uint64 //从系统启动开始累计到当前时刻，nice值为负的进程所占用的CPU时间（单位：jiffies）
	System              uint64 //从系统启动开始累计到当前时刻，内核态时间（单位：jiffies）
	Idle                uint64 //从系统启动开始累计到当前时刻，除硬盘IO等待时间以外其它等待时间（单位：jiffies)
	Iowait              uint64 //从系统启动开始累计到当前时刻，硬盘IO等待时间（单位：jiffies）
	Irq                 uint64 //从系统启动开始累计到当前时刻，硬中断时间（单位：jiffies）
	SoftIrq             uint64 //从系统启动开始累计到当前时刻，软中断时间（单位：jiffies）
	Steal               uint64
	Total               uint64  //user + nice + system + idle + iowait
	IoWaitRate          float64 //io等待时间百分比
	SystemRate          float64 //内核态时间百分比
	UserRate            float64 //用户态时间百分比
	IdleRate            float64 //空闲时间百分比
	ProcsBlocked        uint64  //阻塞进程数
	ProcsRunning        uint64  //运行进程数
	IdleRateSum10       float64 //空闲时间百分比10分钟累加和
	IdleRateSum10Times  int     //空闲时间百分比10分钟累加次数
	IdleRate10          float64 //空闲时间10分钟环比
	IdleRate10Last      int64
	IdleRateSum60       float64 //空闲时间百分比60分钟累加和
	IdleRateSum60Times  int     //空闲时间百分比60分钟累加次数
	IdleRate60          float64 //空闲时间60分钟环比
	IdleRate60Last      int64
	IdleRateSumDay      float64 //空闲时间百分比24h累加和
	IdleRateSumDayTimes int     //空闲时间百分比24h累加次数
	IdleRateDay         float64 //空闲时间日同比
	IdleRateDayLast     int64
}

func (this *Cpu) Dump() {
	fmt.Printf("User:%d, Nice:%d, System:%d, Idle:%d, Iowait:%d, Irq:%d, SoftIrq:%d, Total:%d, IoWaitRate:%f\n",
		this.User,
		this.Nice,
		this.System,
		this.Idle,
		this.Iowait,
		this.Irq,
		this.SoftIrq,
		this.Total,
		this.IoWaitRate)
}

func getWorkTime(t1, t2 CpuTime) uint64 {
	return t1.User - t2.User + t1.Nice - t2.Nice + t1.System - t2.System + t1.Irq - t2.Irq + t1.SoftIrq - t2.SoftIrq + t1.Steal - t2.Steal
}

func getIdleTime(t1, t2 CpuTime) uint64 {
	return t1.Idle - t2.Idle + t1.Iowait - t2.Iowait
}

func (this *Cpu) Collect() error {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return err
	}
	lines := strings.Split(string(contents), "\n")
	if len(lines) == 0 {
		return errors.New("cpu stat err")
	}
	this.UserRate = 0
	this.IdleRate = 0
	for i, line := range lines[1:] {
		if !strings.HasPrefix(line, "cpu") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 11 {
			return errors.New("cpu stat err")
		}

		var cpuTime CpuTime
		cpuTime.User, _ = strconv.ParseUint(fields[1], 10, 64)
		cpuTime.Nice, _ = strconv.ParseUint(fields[2], 10, 64)
		cpuTime.System, _ = strconv.ParseUint(fields[3], 10, 64)
		cpuTime.Idle, _ = strconv.ParseUint(fields[4], 10, 64)
		cpuTime.Iowait, _ = strconv.ParseUint(fields[5], 10, 64)
		cpuTime.Irq, _ = strconv.ParseUint(fields[6], 10, 64)
		cpuTime.SoftIrq, _ = strconv.ParseUint(fields[7], 10, 64)
		if this.Total <= 0 {
			this.CpuTimes = append(this.CpuTimes, cpuTime)
			continue
		}
		workTime := getWorkTime(cpuTime, this.CpuTimes[i])
		idleTime := getIdleTime(cpuTime, this.CpuTimes[i])
		this.UserRate = this.UserRate + 100*float64(workTime)/float64(workTime+idleTime)
		this.IdleRate = this.IdleRate + 100*float64(idleTime)/float64(workTime+idleTime)
		this.CpuTimes[i] = cpuTime
	}
	// this.IdleRate = this.IdleRate / float64(len(this.CpuTimes))
	this.Total = 1
	return nil
}

//io等待时间百分比
func (this *Cpu) IoWaitRateFunc(args string) string {
	return FloatToString(this.IoWaitRate)
}

//内核态时间百分比
func (this *Cpu) SystemRateFunc(args string) string {
	return FloatToString(this.SystemRate)
}

//用户态时间百分比
func (this *Cpu) UserRateFunc(args string) string {
	return FloatToString(this.UserRate)
}

//空闲时间百分比
func (this *Cpu) IdleRateFunc(args string) string {
	//this.AddIdleRate(this.IdleRate)
	return FloatToString(this.IdleRate)
}

//阻塞进程数
func (this *Cpu) ProcsBlockedFunc(args string) string {
	return strconv.FormatUint(this.ProcsBlocked, 10)
}

//运行进程数
func (this *Cpu) ProcsRunningFunc(args string) string {
	return strconv.FormatUint(this.ProcsRunning, 10)
}

/*
//累加空闲时间百分比
func (this *Cpu) AddIdleRate(idleRate float64) {
	this.IdleRateSum10 += idleRate
	this.IdleRateSum10Times++
	this.IdleRateSum60 += idleRate
	this.IdleRateSum60Times++
	this.IdleRateSumDay += idleRate
	this.IdleRateSumDayTimes++
}

func (this *Cpu) ResetIdleRate10() {
	this.IdleRateSum10 = 0
	this.IdleRateSum10Times = 0
}

func (this *Cpu) ResetIdleRate60() {
	this.IdleRateSum60 = 0
	this.IdleRateSum60Times = 0
}

func (this *Cpu) ResetIdleRateDay() {
	this.IdleRateSumDay = 0
	this.IdleRateSumDayTimes = 0
}

//CPU闲置时间10分钟环比
func (this *Cpu) IdleRate10Func(args string) string {
	if time.Now().Unix()-this.IdleRate10Last < 600 {
		return ""
	}
	if this.IdleRateSum10Times <= 0 {
		return ""
	}
	var ret float64 = 0
	avg := this.IdleRateSum10 / float64(this.IdleRateSum10Times)
	if this.IdleRate10 == 0 && avg != 0 {
		ret = 100
	} else {
		ret = (avg - this.IdleRate10) / this.IdleRate10 * 100
	}
	this.IdleRate10 = avg
	this.ResetIdleRate10()
	this.IdleRate10Last = time.Now().Unix()
	return g.FloatToString(ret)
}

//CPU闲置时间小时环比
func (this *Cpu) IdleRate60Func(args string) string {
	if time.Now().Unix()-this.IdleRate60Last < 3600 {
		return ""
	}
	if this.IdleRateSum60Times <= 0 {
		return ""
	}
	var ret float64 = 0
	avg := this.IdleRateSum60 / float64(this.IdleRateSum60Times)
	if this.IdleRate60 == 0 && avg != 0 {
		ret = 100
	} else {
		ret = (avg - this.IdleRate60) / this.IdleRate60 * 100
	}
	this.IdleRate60 = avg
	this.ResetIdleRate60()
	this.IdleRate60Last = time.Now().Unix()
	return g.FloatToString(ret)
}

//cpu闲置时间日同比
func (this *Cpu) IdleRateDayFunc(args string) string {
	if time.Now().Unix()-this.IdleRateDayLast < 86400 {
		return ""
	}
	if this.IdleRateSumDayTimes <= 0 {
		return ""
	}
	var ret float64 = 0
	avg := this.IdleRateSumDay / float64(this.IdleRateSumDayTimes)
	if this.IdleRateDay == 0 && avg != 0 {
		ret = 100
	} else {
		ret = (avg - this.IdleRateDay) / this.IdleRate60 * 100
	}
	this.IdleRateDay = avg
	this.ResetIdleRateDay()
	this.IdleRateDayLast = time.Now().Unix()
	return g.FloatToString(ret)
}*/

//返回某个进程的cpu使用率
func CpuUsedRateByProc(proc string) string {
	//return ExecOutput("ps axuwww|grep " + proc + "|grep -v grep|awk 'BEGIN{sum=0}{sum+=$3 }END{print sum}'")
	return ExecOutput("top -n 1 -b |grep '" + proc + "'|grep -v grep|awk 'BEGIN{sum=0}{sum+=$9 }END{print sum}'")
}
