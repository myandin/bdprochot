package main

import (
        "encoding/binary"
        "fmt"
        "log"
        "syscall"
        "strconv"
)

const defaultFmtStr = "/dev/cpu/%d/msr"

type MSRDev struct {
        fd int
}

func (d MSRDev) Close() error {
        return syscall.Close(d.fd)
}

func (d MSRDev) Read(msr int64) (uint64, error) {
        regBuf := make([]byte, 8)
        rc, err := syscall.Pread(d.fd, regBuf, msr)
        if err != nil {
                return 0, err
        }
        if rc != 8 {
                return 0, fmt.Errorf("Read wrong count of bytes: %d", rc)
        }
        msrOut := binary.LittleEndian.Uint64(regBuf)
        return msrOut, nil
}

func (d MSRDev) Write(regno int64, value uint64) error {
        buf := make([]byte, 8)
        binary.LittleEndian.PutUint64(buf, value)
        count, err := syscall.Pwrite(d.fd, buf, regno)
        if err != nil {
                return err
        }
        if count != 8 {
                return fmt.Errorf("Write count not a uint64: %d", count)
        }
        return nil
}

func WriteMSR(cpu int, msr int64, value uint64) error {
        m, err := MSR(cpu)
        if err != nil {
                return err
        }
        err = m.Write(msr, value)
        if err != nil {
                return err
        }
        return m.Close()
}

func MSR(cpu int) (MSRDev, error) {
        cpuDir := fmt.Sprintf(defaultFmtStr, cpu)
        fd, err := syscall.Open(cpuDir, syscall.O_RDWR, 777)
        if err != nil {
                return MSRDev{}, err
        }
        return MSRDev{fd: fd}, nil
}

func ReadMSR(cpu int, msr int64) (uint64, error) {
        m, err := MSR(cpu)
        if err != nil {
                return 0, err
        }
        msrD, err := m.Read(msr)
        if err != nil {
                return 0, err
        }
        return msrD, m.Close()

}

func main() {
    readdata, err := ReadMSR(0, 0x1FC)
    if err != nil {
        log.Fatalf("Error: %s", err)
    }
    fmt.Println("MSR old value:", readdata)
    mask, err := strconv.ParseUint("FFFFFFFFFE", 16, 64)
    writedata := readdata & mask
    fmt.Println("MSR new value:", writedata)
    err = WriteMSR(0, 0x1FC, writedata)
    if err != nil {
        log.Fatalf("Error: %s", err)
    }
}
