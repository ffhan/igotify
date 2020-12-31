package igotify

/*
#include <unistd.h>

// allows us to use GNU pathconf interface through Go to retrieve NAME_MAX limit for the current system.
long get_name_max() {
	return pathconf(".", _PC_NAME_MAX);
}
*/
import "C"
import "log"

var (
	nameMax = 0 // represents NAME_MAX from GNU sysconfig
)

func init() {
	// set nameMax to value of cgo get_name_max() resulting value
	nameMax = int(C.get_name_max())
	if nameMax <= 0 {
		log.Println("get_name_max() failed - assuming NAME_MAX is 255")
		nameMax = 255
	}
}

func GetNameMax() int {
	return nameMax
}
