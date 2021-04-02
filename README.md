# timewheel
简易时间轮实现延迟任务


### 大致思路

    https://github.com/zlbonly/timewheel/blob/master/pics/timewheel.jpeg
    
    #### 1、timewheel  维护 slots[] 槽位（每个槽内存放 延时任务的双向链表），slotNum  时间轮的槽位数量，interva了
    时间轮中指针每隔多久往前移动一格， 然后根据 Task 维护任务的延迟时间 dealy ， 计算 在时间轮中的位置pos，和圈数
    delaySeconds := int(d.Seconds())  // 延时时间
    intervalSeconds := int(tw.interval.Seconds()) // 移动步长
    cycle = int(delaySeconds / intervalSeconds / tw.slotNum)  // 圈数
    pos = int(tw.currentPos+delaySeconds/intervalSeconds) % tw.slotNum  // 时间轮中所在位置
    
    然后根据 定时器轮询 获取相应槽位，并遍历对应的 任务链表，取出具体任务 执行。
    
    
    #### 缺点：
    1、缺点： 时间轮没有分级，会导致单个槽点 任务双向链表过长，时间轮空转，没有分级概念