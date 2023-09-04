package znet

import (
	"fmt"
	"strconv"
	"zinx/zinx/utils"
	"zinx/zinx/ziface"
)

/*
消息处理模块的实现
*/

type MsgHandle struct {
	//存放每个MsgID 所对应的处理方法
	Apis map[uint32]ziface.IRouter

	//负责worker取任务的消息队列
	TaskQueue []chan ziface.IRequest
	//当前worker工作池的数量
	WorkerPoolSize uint32
}

// 初始化/创建MsgHandle方法
func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		TaskQueue:      make([]chan ziface.IRequest, utils.GlobalObject.WorkerPoolSize),
		WorkerPoolSize: utils.GlobalObject.WorkerPoolSize, //从全局配置中获取

	}
}

// 34 0318
// 调度/执行对应的Router消息处理方法
func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) {

	//从request中找到msgid
	handler, ok := mh.Apis[request.GetMsgId()]

	if !ok {
		fmt.Println("api MsgID = ", request.GetMsgId(), " is not found! Need register")
		return
	}
	//根据msgid 调度对应的router业务即可
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)

}

// 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgID uint32, router ziface.IRouter) {
	//1判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgID]; ok {
		//id已经注册了
		panic("repeat api,msgid = " + strconv.Itoa(int(msgID)))
	}
	//2添加msg与API的绑定关系
	mh.Apis[msgID] = router
	fmt.Println("Add api MsgID = ", msgID, " succ!")
}

// 启动一个worker工作池(开启工作池的动作只能发生一次)
func (mh *MsgHandle) StartWorkPool() {
	//根据workpoolsize分别开启worker,每个worker用go承载
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		//一个worker被启动
		//1 当前的Worker对应的channel消息队列 开辟空间 第i个worker  就用第0个channel
		mh.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalObject.MaxWorkerTaskLen)
		//2 启动当前的worker，阻塞等待消息从channel传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}

}

// 启动一个worker工作流程
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is started ...")
	//不断的阻塞等待对应消息队列的消息
	for {
		select {
		//如果有消息过来，出列的即使一个客户端的Requset，执行当前的Requset所绑定的业务
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// 将消息交给TaskQueue，由Worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	//1 将消息平均分配给不同的worker
	//跟据客户端建立的ConnID来进行分配
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	fmt.Println("Add ConnID = ", request.GetConnection().GetConnID(),
		" request MsgID = ", request.GetMsgId(), " to WorkerID = ", workerID)
	//2 将消息发送给对应的worker的TaskQueue即可
	mh.TaskQueue[workerID] <- request
}
