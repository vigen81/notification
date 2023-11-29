package main

import (
	`gitlab.com/healthcare-integration/stream`
	`go-micro.dev/v4/logger`
	
	pb `gitlab.com/healthcare-integration/golang/notification-service/pb/notification`
)

func main() {
	
	st := stream.New(
		stream.WithAddress("localhost:4222"),
	)
	err := st.Init()
	if nil != err {
		logger.Fatal(err)
	}
	pub := st.Publisher()
	err = pub.PublishInterface("ns.email", &pb.EmailRequest{
		Address:  "vahanerevan1@gmail.com",
		Name:     "vahan",
		Body:     "vahan test",
		Subject:  "vahan test",
		Meta:     nil,
		Schedule: nil,
	})
	if nil != err {
		logger.Fatal(err)
	}
}
