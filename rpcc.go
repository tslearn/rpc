package rpcc

import "github.com/tslearn/rpcc/util"

func make() {
	assert := util.NewAssert(nil)
	assert(1).Equals(6)
}
