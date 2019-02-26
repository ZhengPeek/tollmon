package monitor

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"strings"
	"tollsys/tollmon/datastruct"
	"tollsys/tollmon/g"
	"tollsys/tollmon/h"
	"tollsys/tollmon/parameters"

	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type Message struct {
	MC int
	MT int
	MB []byte
}

var (
	STX byte = 0x02 //Message Start
	ETX byte = 0x03 //Message End

	LenMc int = 2
	LenMT     = 2

	LenTime    = 14
	LenLaneID  = 26
	LenShift   = 2
	LenEmpID   = 4
	LenClass   = 2
	LenType    = 2
	LenPass    = 4
	LenLoan    = 4
	LenForfeit = 4
	LenETCCar  = 1
	LenEmpName = 20

	LenStatus       = 2
	LenOffSet       = 2
	LenThreshold    = 8
	LenCurrent      = 8
	LenCardType     = 2
	LenMode         = 2
	LenPrintTimes   = 2
	LenPrintNoteNo  = 30
	LenFlow         = 4
	LenETCErrorNote = 30
)

const (
	McData  = 0x01 //Data Message Catalog
	McAlert = 0x20 //Alert Message Catalog
	McTest  = 0x30 //Test Message Catalog

	MtEntryLane  = 0x10 //Entry Message Type
	MtExitLane   = 0x11 //Exit
	MtOnduty     = 0x12 //Onduty
	MtEnOffduty  = 0x13 //EnOffDuty
	MtExOffduty  = 0x14 //ExOffDuty
	MtImage      = 0x15 //TranImage
	MtVoice      = 0x16 //Voice
	MtLaneStatus = 0x17 //LaneStatus
	MtGJC        = 0x18 //GJC
	MtReq        = 0x19 //Entry Search Request

	MtClassChange     = 0x01 //Vehicle Class Changed
	MtVio             = 0x02 //Vehicle Vio
	MtDutyEnd         = 0x03 //OffDuty
	MtTypeChange      = 0x04 //Vehicle Type Changed
	MtEntryCard       = 0x05 //Entry Card Storage Min
	MtExitCard        = 0x06 //Exit Card Storage Max
	MtNotePrint       = 0x07 //Print Note
	MtNoteHand        = 0x08 //hand Note
	MtVehicleCount    = 0x09 //loop count
	MtOpeCardFail     = 0x0A //Operate Card Fail
	MtReaderInitFail  = 0x0B //Init Reader Fail
	MtCardModeChange  = 0x0C //Change Send Card Mode
	MtNoteModeChange  = 0x0D //Change Send Note Mode
	MtNoteAgain       = 0x0E //Note Again
	MtExBadCard       = 0x0F //Exit Bad Card
	MtExNoCard        = 0x10 //Exit No Card
	MtSimulate        = 0x11 //Simulate
	MtDebt            = 0x12 //Debt
	MtFree            = 0x13 //Free Car
	MtFlowChange      = 0x14 //change record
	MtMotoStart       = 0x15 //moto start
	MtMotoEnd         = 0x16 //moto end
	MtExitChangeClass = 0x17 //exit change vehicle class
	MtReaderErr       = 0x18 //Reader Error
	MtUType           = 0x19 //UType CAR
	MtOverTime        = 0x20 //OverTime Car
	MtManualAlert     = 0x21 //Manual Alert
	MtETCInfo         = 0x22 //ETC INFO

	MtHeart = 0x22 //hb
)

//报文处理方法
//根据报文头尾定义截获报文并处理
//当截获stop信号量时终止该goroutine
func parseMsg(c chan byte, stop chan int) {
	for {
		select {
		case b := <-c:
			var buffer []byte
			//g.LogDebug("parse msg start...")
			if b == STX {
				//g.LogDebug("STX GET - Buffer Start...")
				buffer = append(buffer, b)
				for {
					v := <-c
					if v == ETX {
						//g.LogDebug("ETX Get - Buffer End...")
						buffer = append(buffer, v)
						handleMsg(buffer)
						break
					}
					buffer = append(buffer, v)
				}
			}
		case <-stop:
			g.LogInfo("receive stop signal - goroutine parseMsg stop")
			close(c)
			return
		}
	}
}

//报文解码方法，根据协议解码
func handleMsg(b []byte) {
	var msg Message
	//s := string(b)
	//g.LogDebug("handle Msg :", s)
	msg.MC = bytesToInt(b[1:3])
	msg.MT = bytesToInt(b[3:5])
	msg.MB = b[5:len(b)]
	switch msg.MC {
	case McAlert:
		func() {
			switch msg.MT {
			case MtClassChange:
				handleMtClassChange(msg.MB)
			case MtVio:
				handleMtVio(msg.MB)
			case MtDutyEnd:
				handleMtDutyEnd(msg.MB)
			case MtTypeChange:
				handleMtTypeChange(msg.MB)
			case MtEntryCard:
				handleMtEntryCard(msg.MB)
			case MtExitCard:
				handleMtExitCard(msg.MB)
			case MtNotePrint:
				handleMtNotePrint(msg.MB)
			case MtNoteHand:
				handleMtNoteHand(msg.MB)
			case MtVehicleCount:
				handleMtVehicleCount(msg.MB)
			case MtOpeCardFail:
				handleMtOpeCardFail(msg.MB)
			case MtReaderInitFail:
				handleMtReaderInitFail(msg.MB)
			case MtCardModeChange:
				handleMtCardModeChange(msg.MB)
			case MtNoteModeChange:
				handleMtNoteModeChange(msg.MB)
			case MtNoteAgain:
				handleMtNoteAgain(msg.MB)
			case MtExBadCard:
				handleMtExBadCard(msg.MB)
			case MtExNoCard:
				handleMtExNoCard(msg.MB)
			case MtSimulate:
				handleMtSimulate(msg.MB)
			case MtDebt:
				handleMtDebt(msg.MB)
			case MtFree:
				handleMtFree(msg.MB)
			case MtFlowChange:
				handleMtFlowChange(msg.MB)
			case MtMotoStart:
				handleMtMotoStart(msg.MB)
			case MtMotoEnd:
				handleMtMotoEnd(msg.MB)
			case MtExitChangeClass:
				handleMtExitChangeClass(msg.MB)
			case MtReaderErr:
				handleMtReaderErr(msg.MB)
			case MtUType:
				handleMtUType(msg.MB)
			case MtOverTime:
				handleMtOverTime(msg.MB)
			case MtManualAlert:
				handleMtManualAlert(msg.MB)
			case MtETCInfo:
				handleMtETCInfo(msg.MB)
			default:
				g.LogError("Unkown Message Type:", msg.MB)
			}
		}()
	case McTest:
		func() {
			//g.LogDebug("handle MT:22 Heart")
			index := 0
			sTime, err := parseTimeFormat(string(subBytes(msg.MB, index, LenTime)))
			index += LenTime
			if err != nil {
				g.LogError("parse Time Err:", err.Error())
			}
			sLaneID := string(subBytes(msg.MB, index, LenLaneID))
			commTime := parseTime(sTime)
			lock.Lock()
			parameters.GetLaneQueue()[sLaneID] = commTime
			lock.Unlock()
			//checkLaneStatus(sLaneID, sTime)
			g.LogDebug("心跳-[Time:", sTime, " - LaneID:", sLaneID, "]")
		}()
	case McData:
		func() {
			switch msg.MT {
			case MtEntryLane:
				handleMtEntryLane(msg.MB)
			case MtExitLane:
				handleMtExitLane(msg.MB)
			case MtOnduty:
				handleMtOnduty(msg.MB)
			case MtEnOffduty:
				handleMtEnOffduty(msg.MB)
			case MtExOffduty:
				handleMtExOffduty(msg.MB)
			case MtImage:
				handleMtImage(msg.MB)
			case MtVoice:
				handleMtVoice(msg.MB)
			case MtLaneStatus:
				handleMtLaneStatus(msg.MB)
			case MtGJC:
				handleMtGJC(msg.MB)
			case MtReq:
				handleMtReq(msg.MB)
			}
		}()
	}
}

func handleMtEntryLane(s []byte) {
	//g.LogDebug("handle MT:10 EntryLane")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(s, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(s, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(s, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(s, index, LenEmpID))
	index += LenEmpID
	a["EnClass"] = bytesToInt(subBytes(s, index, LenClass))
	index += LenClass
	a["EnType"] = bytesToInt(subBytes(s, index, LenType))
	index += LenType
	a["ETCCar"] = bytesToInt(subBytes(s, index, LenETCCar))
	msg := setMsgSend(McData, MtEntryLane, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("入口车道记录信息-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtExitLane(s []byte) {
	//g.LogDebug("handle MT:11 ExitLane")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(s, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(s, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(s, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(s, index, LenEmpID))
	index += LenEmpID
	a["ExClass"] = bytesToInt(subBytes(s, index, LenClass))
	index += LenClass
	a["ExType"] = bytesToInt(subBytes(s, index, LenType))
	index += LenType
	a["Pass"] = bytesToInt(subBytes(s, index, LenPass))
	index += LenPass
	a["Loan"] = bytesToInt(subBytes(s, index, LenLoan))
	index += LenLoan
	a["Forfeit"] = bytesToInt(subBytes(s, index, LenForfeit))
	index += LenForfeit
	a["ETCCar"] = bytesToInt(subBytes(s, index, LenETCCar))
	msg := setMsgSend(McData, MtExitLane, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("出口车道记录信息-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtOnduty(s []byte) {
	//g.LogDebug("handle MT:12 OnDuty")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(s, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(s, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(s, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(s, index, LenEmpID))
	index += LenEmpID
	a["EmpName"] = strings.Trim(bytesToChs(subBytes(s, index, LenEmpName)), " ")
	a["onDutyTime"] = sTime
	msg := setMsgSend(McData, MtOnduty, sTime, sLaneID, a)
	parameters.UpdateLaneInfo(sLaneID, "shiftStatus", true)
	parameters.UpdateLaneInfo(sLaneID, "onDutyTime", a["onDutyTime"])
	parameters.UpdateLaneInfo(sLaneID, "shiftNo", a["Shift"])
	parameters.UpdateLaneInfo(sLaneID, "empName", a["EmpName"])
	parameters.UpdateLaneInfo(sLaneID, "empID", a["EmpID"])
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("上班记录信息-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtEnOffduty(s []byte) {
	//g.LogDebug("handle MT:13 EnOffDuty")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(s, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(s, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(s, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(s, index, LenEmpID))
	index += LenEmpID
	a["EmpName"] = strings.Trim(bytesToChs(subBytes(s, index, LenEmpName)), " ")
	index += LenEmpName
	sOffDutyTime, err := parseTimeFormat(string(subBytes(s, index, LenTime)))
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	a["OffDutyTime"] = sOffDutyTime
	msg := setMsgSend(McData, MtEnOffduty, sTime, sLaneID, a)
	parameters.UpdateLaneInfo(sLaneID, "shiftStatus", false)
	parameters.UpdateLaneInfo(sLaneID, "offDutyTime", a["OffDutyTime"])
	parameters.UpdateLaneInfo(sLaneID, "shiftNo", a["Shift"])
	parameters.UpdateLaneInfo(sLaneID, "empName", a["EmpName"])
	parameters.UpdateLaneInfo(sLaneID, "empID", a["EmpID"])
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("入口下班记录-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtExOffduty(s []byte) {
	//g.LogDebug("handle MT:14 ExOffDuty")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(s, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(s, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(s, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(s, index, LenEmpID))
	index += LenEmpID
	a["EmpName"] = strings.Trim(bytesToChs(subBytes(s, index, LenEmpName)), " ")
	index += LenEmpName
	sOffDutyTime, err := parseTimeFormat(string(subBytes(s, index, LenTime)))
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	a["OffDutyTime"] = sOffDutyTime
	msg := setMsgSend(McData, MtExOffduty, sTime, sLaneID, a)
	parameters.UpdateLaneInfo(sLaneID, "shiftStatus", false)
	parameters.UpdateLaneInfo(sLaneID, "offDutyTime", a["OffDutyTime"])
	parameters.UpdateLaneInfo(sLaneID, "shiftNo", a["Shift"])
	parameters.UpdateLaneInfo(sLaneID, "empName", a["EmpName"])
	parameters.UpdateLaneInfo(sLaneID, "empID", a["EmpID"])
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("出口下班记录-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtImage(s []byte) {
	//g.LogDebug("handle MT:15 Voice")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(s, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(s, index, LenLaneID))
	msg := setMsgSend(McData, MtImage, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("车道图像-[Time:", sTime, " LaneID:", sLaneID, "]")
}
func handleMtVoice(s []byte) {
	//g.LogDebug("handle MT:16 Voice")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(s, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(s, index, LenLaneID))
	msg := setMsgSend(McData, MtVoice, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("语音信息-[Time:", sTime, " LaneID:", sLaneID, "]")
}
func handleMtLaneStatus(s []byte) {
	//g.LogDebug("handle MT:17 Lane Status")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(s, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(s, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(s, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(s, index, LenEmpID))
	index += LenEmpID
	t := bytesToInt(subBytes(s, index, LenStatus))
	msg := setMsgSend(McData, MtLaneStatus, sTime, sLaneID, a)
	switch t {
	case 2:
		parameters.UpdateLaneInfo(sLaneID, "laneStatus", 0)
		a["Status"] = 0
	case 3:
		parameters.UpdateLaneInfo(sLaneID, "laneStatus", 1)
		a["Status"] = 1
	}
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("车道状态-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtGJC(s []byte) {
	//g.LogDebug("handle MT:18 GJC")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(s, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(s, index, LenLaneID))
	msg := setMsgSend(McData, MtGJC, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("代金卡-[Time:", sTime, " LaneID:", sLaneID, "]")
}
func handleMtReq(s []byte) {
	//g.LogDebug("handle MT:19 Req")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(s, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(s, index, LenLaneID))
	msg := setMsgSend(McData, MtReq, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("入口查询请求-[Time:", sTime, " LaneID:", sLaneID, "]")
}
func subStr(src string, index int, length int) string {
	rs := []rune(src)
	rl := len(rs)
	end := 0

	if index < 0 {
		index = rl - 1 + index
	}
	end = index + length

	if index > end {
		index, end = end, index
	}

	if index < 0 {
		index = 0
	}
	if index > rl {
		index = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(rs[index:end])
}

func handleMtClassChange(mb []byte) {
	//g.LogDebug("handle MT:01 ClassChange")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["EnClass"] = bytesToInt(subBytes(mb, index, LenClass))
	index += LenClass
	a["ExPreClass"] = bytesToInt(subBytes(mb, index, LenClass))
	index += LenClass
	a["ExClass"] = bytesToInt(subBytes(mb, index, LenClass))
	msg := setMsgSend(McAlert, MtClassChange, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("出入口车型不一致-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtVio(mb []byte) {
	//g.LogDebug("handle MT:02 VIO")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	msg := setMsgSend(McAlert, MtVio, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("闯关-[Time:", sTime, " LaneID:", sLaneID, " EmpID:", a, "]")
}
func handleMtDutyEnd(mb []byte) {
	//g.LogDebug("handle MT:03 DutyEnd")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["Offset"] = bytesToInt(subBytes(mb, index, LenOffSet))
	msg := setMsgSend(McAlert, MtDutyEnd, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("下班通知-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtTypeChange(mb []byte) {
	//g.LogDebug("handle MT:04 Type Change")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["EnType"] = bytesToInt(subBytes(mb, index, LenType))
	index += LenType
	a["ExType"] = bytesToInt(subBytes(mb, index, LenType))
	msg := setMsgSend(McAlert, MtTypeChange, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("车种不一致-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtEntryCard(mb []byte) {
	//g.LogDebug("handle MT:05 Entry Card")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["Threshold"] = bytesToInt(subBytes(mb, index, LenThreshold))
	index += LenThreshold
	a["Current"] = bytesToInt(subBytes(mb, index, LenCurrent))
	msg := setMsgSend(McAlert, MtEntryCard, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("入口通行卡存量报警-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtExitCard(mb []byte) {
	//g.LogDebug("handle MT:06 Exit Card")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["Threshold"] = bytesToInt(subBytes(mb, index, LenThreshold))
	index += LenThreshold
	a["Current"] = bytesToInt(subBytes(mb, index, LenCurrent))
	msg := setMsgSend(McAlert, MtExitCard, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("出口通行卡存量报警-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtNotePrint(mb []byte) {
	//g.LogDebug("handle MT:07 Note Print")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["Threshold"] = bytesToInt(subBytes(mb, index, LenThreshold))
	index += LenThreshold
	a["Current"] = bytesToInt(subBytes(mb, index, LenCurrent))
	msg := setMsgSend(McAlert, MtNotePrint, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("出口打印票存量报警-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtNoteHand(mb []byte) {
	//g.LogDebug("handle MT:08 Note Hand")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["Threshold"] = bytesToInt(subBytes(mb, index, LenThreshold))
	index += LenThreshold
	a["Current"] = bytesToInt(subBytes(mb, index, LenCurrent))
	msg := setMsgSend(McAlert, MtNoteHand, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("出口定额票余额报警-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtVehicleCount(mb []byte) {
	//g.LogDebug("handle MT:09 Vehicle Count")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	msg := setMsgSend(McAlert, MtVehicleCount, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("车道流量计数-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtOpeCardFail(mb []byte) {
	//g.LogDebug("handle MT:0A Operate Card Failed")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["CardType"] = bytesToInt(subBytes(mb, index, LenCardType))
	msg := setMsgSend(McAlert, MtOpeCardFail, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("卡操作失败-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtReaderInitFail(mb []byte) {
	//g.LogDebug("handle MT:0B Reader Init Fail")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	msg := setMsgSend(McAlert, MtReaderInitFail, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("卡机初始化失败-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtCardModeChange(mb []byte) {
	//g.LogDebug("handle MT:0C Card Mode Changed")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["OrigMode"] = bytesToInt(subBytes(mb, index, LenMode))
	index += LenMode
	a["CurrMode"] = bytesToInt(subBytes(mb, index, LenMode))
	msg := setMsgSend(McAlert, MtCardModeChange, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("入口发卡模式改变-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtNoteModeChange(mb []byte) {
	//g.LogDebug("handle MT:0D Note Mode Changed")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["OrigMode"] = bytesToInt(subBytes(mb, index, LenMode))
	index += LenMode
	a["CurrMode"] = bytesToInt(subBytes(mb, index, LenMode))
	msg := setMsgSend(McAlert, MtNoteModeChange, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("票据模式改变-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtNoteAgain(mb []byte) {
	//g.LogDebug("handle MT:0E Note Again")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["PrintTimes"] = bytesToInt(subBytes(mb, index, LenPrintTimes))
	index += LenPrintTimes
	a["PrintNoteNo"] = string(subBytes(mb, index, LenPrintNoteNo))
	msg := setMsgSend(McAlert, MtNoteAgain, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("发票重打-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtExBadCard(mb []byte) {
	//g.LogDebug("handle MT:0F Exit Bad Card")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["Class"] = bytesToInt(subBytes(mb, index, LenClass))
	index += LenClass
	a["Type"] = bytesToInt(subBytes(mb, index, LenType))
	msg := setMsgSend(McAlert, MtExBadCard, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("出口坏卡-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtExNoCard(mb []byte) {
	//g.LogDebug("handle MT:10 Exit No Card")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["Class"] = bytesToInt(subBytes(mb, index, LenClass))
	index += LenClass
	a["Type"] = bytesToInt(subBytes(mb, index, LenType))
	msg := setMsgSend(McAlert, MtExNoCard, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("出口无卡-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtSimulate(mb []byte) {
	//g.LogDebug("handle MT:11 Simulate")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	msg := setMsgSend(McAlert, MtSimulate, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("模拟放车-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtDebt(mb []byte) {
	//g.LogDebug("handle MT:12 Debt")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["Class"] = bytesToInt(subBytes(mb, index, LenClass))
	index += LenClass
	a["Type"] = bytesToInt(subBytes(mb, index, LenType))
	msg := setMsgSend(McAlert, MtDebt, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("欠款未付车辆-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtFree(mb []byte) {
	//g.LogDebug("handle MT:13 Free")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["Type"] = bytesToInt(subBytes(mb, index, LenType))
	index += LenType
	msg := setMsgSend(McAlert, MtFree, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("免费车辆-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtFlowChange(mb []byte) {
	//g.LogDebug("handle MT:14 Flow Change")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	msg := setMsgSend(McAlert, MtFlowChange, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("流水修改-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtMotoStart(mb []byte) {
	//g.LogDebug("handle MT:15 Moto End")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	msg := setMsgSend(McAlert, MtMotoStart, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("车队开始-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}

func handleMtMotoEnd(mb []byte) {
	//g.LogDebug("handle MT:16 Moto End")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["Flow"] = bytesToInt(subBytes(mb, index, LenFlow))
	msg := setMsgSend(McAlert, MtMotoEnd, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("车队结束-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtExitChangeClass(mb []byte) {
	//g.LogDebug("handle MT:17 Exit Change Class")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["ExPreClass"] = bytesToInt(subBytes(mb, index, LenClass))
	index += LenClass
	a["ExClass"] = bytesToInt(subBytes(mb, index, LenClass))
	msg := setMsgSend(McAlert, MtExitChangeClass, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("出口车型修改-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtReaderErr(mb []byte) {
	//g.LogDebug("handle MT:18 Reader Err")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	msg := setMsgSend(McAlert, MtReaderErr, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("卡机故障-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtUType(mb []byte) {
	//g.LogDebug("handle MT:19 UType")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["ExClass"] = bytesToInt(subBytes(mb, index, LenClass))
	msg := setMsgSend(McAlert, MtUType, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("U行车-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtOverTime(mb []byte) {
	//g.LogDebug("handle MT:20 Over Time")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	index += LenEmpID
	a["ExClass"] = bytesToInt(subBytes(mb, index, LenClass))
	msg := setMsgSend(McAlert, MtOverTime, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("超时车辆-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtManualAlert(mb []byte) {
	//g.LogDebug("handle MT:21 Manual Alert")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["Shift"] = bytesToInt(subBytes(mb, index, LenShift))
	index += LenShift
	a["EmpID"] = bytesToInt(subBytes(mb, index, LenEmpID))
	msg := setMsgSend(McAlert, MtManualAlert, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("人工报警-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func handleMtETCInfo(mb []byte) {
	//g.LogDebug("handle MT:22 ETC Info")
	a := make(map[string]interface{})
	index := 0
	sTime, err := parseTimeFormat(string(subBytes(mb, index, LenTime)))
	index += LenTime
	if err != nil {
		g.LogError("parse Time Err:", err.Error())
	}
	sLaneID := string(subBytes(mb, index, LenLaneID))
	index += LenLaneID
	a["ETCErrorNote"] = string(subBytes(mb, index, LenETCErrorNote))
	msg := setMsgSend(McAlert, MtETCInfo, sTime, sLaneID, a)
	h.PushRealData(sLaneID[0:16], msg)
	g.LogDebug("ETC信息-[Time:", sTime, " LaneID:", sLaneID, a, "]")
}
func parseTimeFormat(time string) (string, error) {
	if len(time) != 14 {
		g.LogError("vaild time:", time)
		err := errors.New("vaild time")
		return "", err
	}
	var buffer bytes.Buffer
	buffer.WriteString(subStr(time, 0, 4))
	buffer.WriteString("-")
	buffer.WriteString(subStr(time, 4, 2))
	buffer.WriteString("-")
	buffer.WriteString(subStr(time, 6, 2))
	buffer.WriteString(" ")
	buffer.WriteString(subStr(time, 8, 2))
	buffer.WriteString(":")
	buffer.WriteString(subStr(time, 10, 2))
	buffer.WriteString(":")
	buffer.WriteString(subStr(time, 12, 2))
	return buffer.String(), nil
}
func subBytes(r []byte, index int, l int) []byte {
	lr := len(r)
	if index+l < lr {
		return r[index : index+l]
	}
	return r[index:]
}
func bytesToInt(b []byte) int {
	bs, _ := hex.DecodeString(string(b))
	//bs, _ := hex.DecodeString(string(b))
	if len(bs) == 1 {
		bs = append([]byte{0x00, 0x00, 0x00}, bs[0])
	}
	if len(bs) == 2 {
		bs = append([]byte{0x00, 0x00}, bs[0], bs[1])
	}
	if len(bs) == 3 {
		bs = append([]byte{0x00}, bs[0], bs[1], bs[2])
	}
	if len(bs) > 4 {
		g.LogError("err length ", len(bs))
	}
	bytesBuffer := bytes.NewBuffer(bs)
	var tmp int32
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int(tmp)
}
func bytesToChs(gbkBs []byte) string {
	s := string(gbkBs)
	gs, _ := hex.DecodeString(s)
	i := bytes.NewBuffer(gs)
	o := transform.NewReader(i, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(o)
	if e != nil {
		g.LogError("hexStringToChs err", e.Error())
	}
	t := string(d)
	return t
}
func hexStringToChs(s string) string {
	gbkBs, _ := hex.DecodeString(s)
	i := bytes.NewBuffer(gbkBs)
	o := transform.NewReader(i, simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(o)
	if e != nil {
		g.LogError("hexStringToChs err", e.Error())
	}
	t := string(d)
	return t
}
func setMsgSend(mc int, mt int, mTime string, mLane string, a map[string]interface{}) datastruct.MsgSend {
	//TODO 发布前需确认测试数据已注释
	msg := datastruct.NewMsgSend()
	msg.MsgCatalog = mc
	msg.MsgType = mt
	msg.MsgTime = mTime
	msg.MsgLane = mLane
	msg.MsgContent = a

	//if g.Config().Log.Debug {
	//	msg.MsgCatalog = 32
	//	testMap := make(map[string]interface{})
	//	testMap["shift"] = 1
	//	testMap["empID"] = 1234
	//	msg.MsgContent = testMap
	//	msg.MsgType = 2
	//}

	return msg
}
func parseTime(timeStr string) time.Time {
	result, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr, time.Local)
	return result
}
