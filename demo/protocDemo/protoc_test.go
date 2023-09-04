package protocDemo

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"testing"
	"zinx/demo/protocDemo/pb"
)

func TestProtoc(t *testing.T) {
	person := &pb.Person{
		Name:   "zhangdayu",
		Age:    22,
		Emails: []string{"163.com", "gmail.com", "qq.com"},
		Phones: []*pb.PhoneNumber{
			&pb.PhoneNumber{
				Number: "15510113599",
				Type:   pb.PhoneType_MOBILE,
			},
			&pb.PhoneNumber{
				Number: "18511108847",
				Type:   pb.PhoneType_HOME,
			},
			&pb.PhoneNumber{
				Number: "13031054227",
				Type:   pb.PhoneType_WORK,
			},
		},
	}
	//将person对象 protobuf的message进行序列化，得到二进制文件
	data, err := proto.Marshal(person)
	//data 就是要传输的数据  对端需要按照message Person格式进行解析
	if err != nil {
		fmt.Println("marshal err ", err)
	}
	//解码
	newdata := &pb.Person{}
	err = proto.Unmarshal(data, newdata)
	if err != nil {
		fmt.Println("Unmarshal err ", err)
	}
	fmt.Println("源数据 ", person)
	fmt.Println("编码后数据 ", newdata)
}
