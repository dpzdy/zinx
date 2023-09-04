package znet

import (
	"fmt"
	"net"
	"zinx/zinx/utils"
	"zinx/zinx/ziface"
)

// IServer的接口实现，定义一个Server的服务器模块
type Server struct {
	//服务器名称
	Name string
	//服务器绑定的IP版本
	IPVersion string
	//服务器监听的IP
	IP string
	//服务器监听的端口
	Port int
	//当前server的消息管理模块，用来绑定MSgID和对应的处理业务api关系
	MsgHandler ziface.IMsgHandle
	//该server的链接管理器
	ConnMgr ziface.IConnManager
	//该server创建链接之后自动调用的Hook函数
	OnConnStart func(conn ziface.IConnection)
	//该server销毁链接之后自动调用的Hook函数
	OnConnStop func(conn ziface.IConnection)
}

// 创建addr listenner 处理客户端业务
func (s *Server) Start() {
	fmt.Printf("[Zinx] Server Name : %s, listenner at IP : %s, Port : %d is starting\n",
		utils.GlobalObject.Name, utils.GlobalObject.Host, utils.GlobalObject.TcpPort)
	fmt.Printf("[Zinx] Version %s, MaxConn : %d, MaxPackageSize : %d\n",
		utils.GlobalObject.Version, utils.GlobalObject.MaxConn, utils.GlobalObject.MaxPackageSize)

	go func() {
		//0 开启消息队列及worker工作池
		s.MsgHandler.StartWorkPool()
		//1 获取一个tcp的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr error : ", err)
			return
		}
		//2 监听服务器的地址
		listenner, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen ", s.IPVersion, " err", err)
			return
		}
		fmt.Println("start Zinx server succ, ", s.Name, " succ, Listening...")
		var cid uint32
		cid = 0
		//3 阻塞的等待客户端链接，处理客户端链接业务（读写）
		for {
			//如果有客户端链接进来，阻塞会返回
			conn, err := listenner.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err ", err)
				continue
			}

			//设置最大链接个数的判断，若超过，关闭新链接
			if s.ConnMgr.Len() >= utils.GlobalObject.MaxConn {
				//TODO 给客户端相应一个超出最大链接的错误包
				fmt.Println("Too Many Connection MacConn = ", utils.GlobalObject.MaxConn)
				conn.Close()
				continue
			}
			//将处理新连接的业务方法 和 conn 进行绑定  得到链接模块
			dealConn := NewConnection(s, conn, cid, s.MsgHandler)
			cid++
			//启动当前的链接业务处理
			go dealConn.Start()

		}
	}()
}

// 43 2
func (s *Server) Stop() {
	//将一些服务器资源、状态、连接信息  进行停止回收
	fmt.Println("[STOP] Zinx Server name ", s.Name)
	s.ConnMgr.ClearConn()
}

// 运行服务器，调用Start()方法，调用后做阻塞处理，在之间做一个扩展功能
func (s *Server) Serve() {
	//启动server服务功能
	s.Start()

	//TODO 做一些服务器之后的额外业务

	//阻塞状态
	select {}
}

// 路由功能：给当前服务注册一个路由方法，供客户端的链接处理使用
func (s *Server) AddRouter(msgID uint32, router ziface.IRouter) {
	s.MsgHandler.AddRouter(msgID, router)
	fmt.Println("Add router Succ!")
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

/*
初始化Server模块的方法,返回抽象层
*/
func NewServer(name string) ziface.IServer {
	s := &Server{
		Name:       utils.GlobalObject.Name,
		IPVersion:  "tcp4",
		IP:         utils.GlobalObject.Host,
		Port:       utils.GlobalObject.TcpPort,
		MsgHandler: NewMsgHandle(),
		ConnMgr:    NewConnManager(),
	}
	return s
}

// 注册OnConnStart钩子函数方法
func (s *Server) SetOnConnStart(hookFunc func(conn ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

// 注册OnConnStop钩子函数方法
func (s *Server) SetOnConnStop(hookFunc func(conn ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// 调用OnConnStart钩子函数方法
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("-----> Call On StartFunc")
		s.OnConnStart(conn)
	}
}

// 调用OnConnStop钩子函数方法
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("-----> Call On StopFunc")
		s.OnConnStop(conn)
	}
}
