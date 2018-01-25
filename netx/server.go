package netx

import (
	"net"
	"strconv"
	"strings"
	"bytes"
	"time"
	"log"
)

/**
存储链接中的客户端
 */
var client_nick_2_conn_map = make(map[string]net.Conn)
var client_conn_2_nick_map = make(map[net.Conn]string)


var user_msg_chan = make(chan *user_msg)

type user_msg struct {
	Nick string
	Msg string
}

func StartServer(port int){
	service := string(strconv.AppendInt([]byte(":") , int64(port) , 10))
	tcpAddr , err := net.ResolveTCPAddr("tcp4" , service)
	if err != nil {
		log.Fatalln("open server error " , err)
	}
	listen , err := net.ListenTCP("tcp" , tcpAddr)
	if err != nil {
		log.Fatalln("listen server error " , err)
	}
	go pushMsgJob()
	log.Println("start server "+ service +" ok")
	for {
		conn , err := listen.Accept()
		if err != nil {
			log.Println("accept client error " , err)
			continue
		}
		log.Println("accept client ok ,client:" , conn)
		go handleClient(conn)
	}
}

func pushMsgJob(){
	for {
		msg := <-user_msg_chan
		for nick , conn := range client_nick_2_conn_map {
			if nick != msg.Nick {
				writeLine(conn , msg.Msg)
			}
		}
	}
}

func handleClient(conn net.Conn){
	defer conn.Close()
	printWelcome(conn)
	writeLine(conn , "请输入昵称:")
	nick , err := choseNick(conn)
	if err != nil {
		log.Println("chose nick error , " , err)
		return
	}
	client_nick_2_conn_map[nick] = conn
	client_conn_2_nick_map[conn] = nick
	defer func(){
		publishMsg(nick , "退出聊天室")
		delete(client_nick_2_conn_map , nick)
		delete(client_conn_2_nick_map , conn)
	}()
	publishMsg(nick , "进入聊天室")
	for {
		buf := make([]byte , 128)
		len , err := conn.Read(buf)
		if err != nil {
			log.Println("read data error , conn:" , conn , " err:" , err)
			break
		}
		msg := strings.TrimSpace(string(buf[:len]))
		if strings.TrimSpace(msg) == "" {
			continue
		}
		nick := client_conn_2_nick_map[conn]
		if msg == "exit" {
			break
		}else{
			publishMsg(nick , msg)
		}
	}
}

func publishMsg(nick string , msg string){
	var buf = bytes.Buffer{}
	buf.WriteString(nick)
	buf.WriteString(" ")
	buf.WriteString(time.Now().Format("15:04:05"))
	buf.WriteString("\r\n")
	buf.WriteString(msg)
	buf.WriteString("\r\n")
	user_msg_chan <- &user_msg{nick , buf.String()}	//推消息道chan
}

func choseNick(conn net.Conn) (string , error){
	for {
		buf := make([]byte , 36)
		len , err := conn.Read(buf)
		if err != nil {
			log.Println("read data error , conn:" , conn , " err:" , err)
			return "" ,err
		}
		nick := strings.TrimSpace(string(buf[:len]))
		if nick == "" {
			continue
		}
		if client_nick_2_conn_map[nick] != nil {	//已存在
			writeLine(conn , "昵称已存在，请重新输入:")
			continue
		}else{
			client_nick_2_conn_map[nick] = conn
			client_conn_2_nick_map[conn] = nick
			writeLine(conn , "恭喜你，进入房间成功，现在可以畅所欲言啦.")
			return nick , nil
		}
	}
}

func printWelcome(conn net.Conn){
	userCount := len(client_nick_2_conn_map)
	if userCount > 0 {
		writeLine(conn , "欢迎来到聊天室，当前在线用户：")
		var buf = bytes.Buffer{}
		for nick := range client_nick_2_conn_map {
			buf.WriteString(nick)
			buf.WriteString(" ")
		}
		writeLine(conn , buf.String())
	}else{
		writeLine(conn , "欢迎来到聊天室，当前还没有用户在线哦")
	}
}

func writeLine(conn net.Conn , msg string) {
	conn.Write([]byte(msg))
	conn.Write([]byte("\r\n"))
}