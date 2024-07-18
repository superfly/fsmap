package fsmap

/*
#include <sys/ioctl.h>
#include <linux/fsmap.h>
*/
import "C"

const FS_IOC_GETFSMAP = C.FS_IOC_GETFSMAP
const FMR_OWN_FREE = C.FMR_OWN_FREE /* free space */

const FMR_OF_SPECIAL_OWNER = C.FMR_OF_SPECIAL_OWNER /* owner is a special value */
const FMR_OF_LAST = C.FMR_OF_LAST                   /* segment is the last in the dataset */

type FSMap C.struct_fsmap
type FSMapHead C.struct_fsmap_head

const Sizeof_FSMap = C.sizeof_struct_fsmap
const Sizeof_FSMapHead = C.sizeof_struct_fsmap_head
const FSMapEntries = 1024
const Sizeof_FSMapEntries = Sizeof_FSMapHead + FSMapEntries*Sizeof_FSMapHead
