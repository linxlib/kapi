package xchan

// XChan 队列的封装
type XChan struct {
	c  chan interface{}
	do func(interface{})
}

// NewXChan 创建新的XChan
func NewXChan(cap int, fn func(interface{})) *XChan {
	xc := new(XChan)
	xc.c = make(chan interface{}, cap)
	xc.do = fn

	go xc.run()

	return xc
}

// Add 新增
func (x *XChan) Add(t interface{}) {
	x.c <- t
}

func (x *XChan) run() {
	defer close(x.c)
	for {
		select {
		case item := <-x.c:
			if x.do != nil {
				x.do(item)
			}
		}
	}
}
