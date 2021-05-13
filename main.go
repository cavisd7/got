package main

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

func main() {

	ptm, pts, err := openPts()
	if err != nil {
		os.Exit(1)
	}
}

/*
 *	Create a pseudoterminal (pty)
 *	A pty consists of a pair of virtual character devices: a master (ptm) and a slave (pts)
 *
 *	@return ptm *os.File(?) - Pointer to pty master file
 *	@return pts *os.File - pointer to pty slave file
 *	@return err error - error
 */
func openPty() (ptm, pts *os.File, err error) {

	/*
	 * 	Opening /dev/ptmx gives:
	 *		A file descriptor (fd) for a pty master (ptm)
	 *		A pty slave (pts) device, created in /dev/pts/
	 */
	ptmf, err := os.OpenFile("/dev/ptmx", os.O_RDWR, 0)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err != nil {
			_ = ptmf.Close()
		}
	}()

	/*
	 *	Find the name of the pts created in /dev/pts/
	 */
	var n uint32 // Device unit number Ex. /dev/pts/n
	// TIOCGPTN - syscall command for obtaining device unit number needed for generating the name ofthe pts
	err := ioctl(ptmf.Fd(), syscall.TIOCGPTN, unitptr(unsafe.Pointer(&n)))
	if err != nil {
		return nil, nil, err
	}
	ptsName := "/dev/pts/" + strconv.Itoa(int(n))

	if err := unlockpt(ptmf); err != nil {
		return nil, nil, err
	}

	// Open pts
	ptsf, err := os.OpenFile(ptsName, os.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return nil, nil, err
	}

	return ptmf, ptsf, nil
}

/*
 *	@param fd - File descriptor of ptm
 *	@param cmd - Syscall command
 *	@param ptr uintptr - out
 */
func ioctl(fd, cmd, ptr uintptr) error {

	// TIOCGPTN - Takes a pointer to an unsigned int and provides the number (device unit number) of the pty
	_, _, e := syscall.Syscall(syscall.SYS_IOCTL, fd, cmd, ptr)
	if e != 0 {
		return e
	}

	return nil
}

/*
 *	Unlock pts device that corresponds to ptm
 *	Should be called before opening slave side of pseudoterminal
 *
 *	@param ptmf *os.File - Pointer to ptm file
 *	@return error - error
 */
func unlockpt(ptmf *os.File) error {

	var u uint32

	// TIOCSPTLCK - set or remove lock on pseudoterminal slave device
	return ioctl(ptmf.Fd(), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&u)))
}
