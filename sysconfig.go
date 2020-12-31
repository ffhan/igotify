package igotify

/*
#include <unistd.h>

// allows us to use GNU pathconf interface through Go to retrieve NAME_MAX limit for the current system.
long get_name_max() {
	return pathconf(".", _PC_NAME_MAX);
}
*/
import "C"

var (
	nameMax = 0 // represents NAME_MAX from GNU sysconfig
)

func init() {
	// set nameMax to value of cgo get_name_max() resulting value
	nameMax = int(C.get_name_max())
	if nameMax <= 0 {
		panic("cannot get NAME_MAX value from the system!")
	}
}

func GetNameMax() int {
	return nameMax
}
