package user

import (
	"errors"
	"fmt"
	"github.com/rpccloud/rpc"
	"github.com/rpccloud/rpc/app/util"
	"math/rand"
	"sync"
	"time"
)

const maxPhoneCheckCodeTimeoutMS = 120000
const minSendInternalMS = 30000

var getCheckManager = func() func() *checkManager {
	gManager := &checkManager{}
	rand.Seed(rpc.TimeNow().UnixNano())

	go func() {
		time.Sleep(10 * time.Second)
		gManager.onTimer()
	}()

	return func() *checkManager {
		return gManager
	}
}()

var checkItemCache = sync.Pool{
	New: func() interface{} {
		return &checkItem{}
	},
}

type checkItem struct {
	code      string
	startMS   int64
	timeoutMS int64
}

func newCheckItem(code string, timeoutMS int64) *checkItem {
	ret := checkItemCache.Get().(*checkItem)
	ret.code = code
	ret.startMS = util.TimeNowMS()
	ret.timeoutMS = timeoutMS
	return ret
}

func (p *checkItem) Release() {
	checkItemCache.Put(p)
}

type checkManager struct {
	phoneMap     map[string]*checkItem
	phoneMapLock sync.Mutex
}

func (p *checkManager) onTimer() {
	p.phoneMapLock.Lock()
	defer p.phoneMapLock.Unlock()
	nowMS := util.TimeNowMS()
	for k, v := range p.phoneMap {
		if nowMS-v.startMS > v.timeoutMS {
			delete(p.phoneMap, k)
			v.Release()
		}
	}
}

func (p *checkManager) SendCheckCode(globalPhone string) error {
	buf := [6]byte{48, 48, 48, 48, 48, 48}

	// write rand code to buf
	randCode := rand.Uint64() % 1000000
	fmt.Println("SendCheckCode ", randCode)
	idx := 5
	for idx >= 0 {
		buf[idx] = byte(48 + randCode%10)
		randCode = randCode / 10
		idx--
	}

	// get send code
	code := string(buf[:])

	// record code
	canSend := func() bool {
		p.phoneMapLock.Lock()
		defer p.phoneMapLock.Unlock()

		if v, ok := p.phoneMap[globalPhone]; !ok {
			p.phoneMap[globalPhone] = newCheckItem(code, maxPhoneCheckCodeTimeoutMS)
			return true
		} else if nowMS := util.TimeNowMS(); nowMS-v.startMS > minSendInternalMS {
			v.code = code
			v.startMS = nowMS
			return true
		} else {
			return false
		}
	}()

	if !canSend {
		return errors.New("limit error")
	}

	fmt.Println("SendCheckCode ", globalPhone, code)
	return nil
}

func (p *checkManager) CheckCode(globalPhone string, code string) bool {
	p.phoneMapLock.Lock()
	defer p.phoneMapLock.Unlock()

	if v, ok := p.phoneMap[globalPhone]; ok && v.code == code {
		delete(p.phoneMap, globalPhone)
		v.Release()
		return true
	} else {
		return false
	}
}
