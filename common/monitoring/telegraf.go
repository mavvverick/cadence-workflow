package monitoring

import (
	"fmt"
	"net"
	"os"
)

//EventMessage is an interface
type EventMessage interface {
	Message() string
}

//AIEvent is a struct for log events of user login
type AIEvent struct {
	PostID  string
	Meta    string
	IsTrue  bool
	Version string
}

//VideoCostEvent is a struct for log events of videos processed
type VideoCostEvent struct {
	Status             string
	TaskToken          string
	Event              string
	PostID             string
	VideoType          string
	VideoDuration      float64
	VideoQuality       string
	TimeCostMultiplier int32
}

//UDPConnection makes a connection to the telegraf server
func UDPConnection() (*net.UDPConn, error) {
	TelegrafAddress, err := net.ResolveUDPAddr("udp", os.Getenv("TELEGRAPH_ADDRESS"))
	if err != nil {
		fmt.Println("Error in UDP telegraf address resolution: ", err)
		return nil, err
	}
	conn, err := net.DialUDP("udp", nil, TelegrafAddress)
	if err != nil {
		fmt.Println("Error in UDP telegraf connection: ", err)
		return nil, err
	}

	return conn, nil
}

//FireEvent sends data to the udp server
func FireEvent(conn *net.UDPConn, message string) {
	_, err := conn.Write([]byte(message))
	if err != nil {
		//TODO: Message wsn't sent, do something
	}
}

//Message construct message
func (aie *AIEvent) Message() string {
	return fmt.Sprintf("ai_events,plat=android,is_true=%v,version=%s meta=\"%s\",post_id=\"%s\"", aie.IsTrue, aie.Version, aie.Meta, aie.PostID)
}

//Message construct message
func (vce *VideoCostEvent) Message() string {
	return fmt.Sprintf("cost_video_events,status=%s,event=%s,video_quality=%s,video_type=%s post_id=\"%s\",task_token=\"%s\",time_mult=%d,video_duration=%.2f", vce.Status, vce.Event, vce.VideoQuality, vce.VideoType, vce.PostID, vce.TaskToken, vce.TimeCostMultiplier, vce.VideoDuration)
}
