package persistence
import(
	"sync"
	"time"
)
type MemoryContainer struct{
	_msg_lock sync.Mutex
	_user_lock sync.Mutex
	// 存储实际消息
	msgMap map[string]*OfflineMsg
	// 存储用户消息关系
	userMsgMap map[string][]string
}


func(this *MemoryContainer) Get(id string) (*OfflineMsg,error){
	this._msg_lock.Lock()
	defer func(){
		this._msg_lock.Unlock()
	}()
	if v,ok:= this.msgMap[id];ok{
		delete(this.msgMap,id)
		return v,nil
	}else{
		return nil,NotFoundMsg
	}
}
func (this *MemoryContainer)GetUserMsg(userID string) ([]*OfflineMsg,error){
	this._user_lock.Lock()
	msgs,ok:=this.userMsgMap[userID]
	delete(this.userMsgMap,userID)
	this._user_lock.Unlock()

	if !ok{
		return nil,NotFoundMsg
	}
	msgColl:=make([]*OfflineMsg,0)
	for _,id:=range msgs{
		if v,err:=this.Get(id);err==nil{
			msgColl= append(msgColl,v)
		}
	}
	if len(msgColl)>0{
		return msgColl,nil
	}else{
		return nil,NotFoundMsg
	}
}
func (this *MemoryContainer)Put(msg *OfflineMsg){
	this._msg_lock.Lock()
	this._user_lock.Lock()
	defer func(){
		this._user_lock.Unlock()
		this._msg_lock.Unlock()
	}()
	user:=msg.To
	if u,ok:=this.userMsgMap[user];ok{
		this.userMsgMap[user]=append(u,msg.MsgId)
	}else{
		this.userMsgMap[user]=[]string{msg.MsgId}
	}
	this.msgMap[msg.MsgId]=msg
}
// 清理过期消息
// param interval 间隔（毫秒）
func (this *MemoryContainer) gc(interval int){
	go func(interval int){
		for{
			invalid:=make([]string,0)
			this._msg_lock.Lock()
			for key:=range this.msgMap{
				if this.msgMap[key].KeepLiveTime.Before(time.Now()){
					invalid= append(invalid,key)
				}
			}
			if len(invalid) > 0{
				for _,k:=range invalid{
					delete(this.msgMap,k)
				}
			}
			this._msg_lock.Unlock()
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}(interval)

}