package timewheel

// @author @憨憨
import (
	"container/list"
	"time"
)

// Job延时任务回调函数
type Job func(interface{})

// 延时任务
type Task struct {
	delay time.Duration		// 延迟时间
	data  interface{}		//
	cycle int				//
	key   interface{}		// 定时器唯一标示，用于删除定时器
}

// TimeWheel时间轮
type TimeWheel struct {
	interval          time.Duration	// 指针每隔多久往前移动一格
	slots             []*list.List	// 时间轮槽
	slotNum           int 	// 槽数量
	currentPos        int	// 当前指针指向的槽数量
	job               Job	// 定时器回调函数
	// key: 定时器唯一标识 value: 定时器所在的槽, 主要用于删除定时器, 不会出现并发读写，不加锁直接访问
	timer             map[interface{}]int
	ticker            *time.Ticker
	addTaskChannel    chan Task	// 新增任务channel
	removeTaskChannel chan interface{}	// 删除任务channel
	stopTaskChannel   chan bool	// 停止定时器channel
}

//	创建时间轮
func New(interval time.Duration, slotNum int, job Job) *TimeWheel {
	if interval < 0 || slotNum < 0 || job == nil {
		return nil
	}
	tw := &TimeWheel{
		interval:          interval,
		slots:             make([]*list.List,slotNum),
		slotNum:           slotNum,
		currentPos:        0,
		job:               job,
		timer:             make(map[interface{}]int),
		addTaskChannel:    make(chan Task),
		removeTaskChannel: make(chan interface{}),
		stopTaskChannel:   make(chan bool),
	}
	tw.InitSlots()
	return tw
}


// 初始化槽，每个槽指向一个双向链表
func (tw *TimeWheel) InitSlots() {
	for i := 0; i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
}

// start 启动时间轮
func (tw *TimeWheel) StartWheel() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.Start()
}

// 停止时间轮
func (tw *TimeWheel) Stop() {
	tw.stopTaskChannel <- true
}

// AddTimer 添加定时器，key为定时唯一标示
func (tw *TimeWheel) AddTimer(delay time.Duration, key interface{}, data interface{}) {
	if delay < 0 {
		return
	}
	tw.addTaskChannel <- Task{delay: delay, key: key, data: data}
}

// RemoveTimer 删除定时器 key为添加定时器时传递的定时器唯一标识

func (tw *TimeWheel) RemoveTimer(key interface{}) {
	if key == nil {
		return
	}
	tw.removeTaskChannel <- key
}

func (tw *TimeWheel) Start() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskChannel:
			tw.AddTask(&task)
		case key := <-tw.removeTaskChannel:
			tw.RemoveTask(key)
		case <-tw.stopTaskChannel:
			tw.ticker.Stop()
		}
	}
}

func (tw *TimeWheel) tickHandler() {
	l := tw.slots[tw.currentPos]
	tw.scanAndRunTask(l)
	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
}

 // 扫描链表中过期定时器, 并执行回调函数
func (tw *TimeWheel) scanAndRunTask(l *list.List) {
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.cycle > 0 {
			task.cycle--
			e = e.Next()
			continue
		}
		go tw.job(task.data)
		next := e.Next()
		l.Remove(e)
		if task.key != nil {
			delete(tw.timer, task.key)
		}
		e = next

	}

}

// 新增任务到链表中
func (tw *TimeWheel) AddTask(task *Task) {
	if task.delay < 0 {
		return
	}
	cycle, pos := tw.getPosAndCycleWheel(task.delay)
	task.cycle = cycle
	tw.slots[pos].PushBack(task)
	if task.key != nil {
		tw.timer[task.key] = pos
	}
}

func (tw *TimeWheel) RemoveTask(key interface{}) {
	pos, ok := tw.timer[key]
	if !ok {
		return
	}

	l := tw.slots[pos]
	for e := l.Front(); e != nil; {
		task := e.Value.(*Task)
		if task.key == key {
			delete(tw.timer, task.key)
			l.Remove(e)
		}
		e = e.Next()
	}
}

// 获取定时器在槽中的位置, 时间轮需要转动的圈数
func (tw *TimeWheel) getPosAndCycleWheel(d time.Duration) (cycle, pos int) {
	delaySeconds := int(d.Seconds())
	intervalSeconds := int(tw.interval.Seconds())
	cycle = int(delaySeconds / intervalSeconds / tw.slotNum)
	pos = int(tw.currentPos+delaySeconds/intervalSeconds) % tw.slotNum
	return cycle, pos
}
