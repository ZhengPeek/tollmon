package t

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"tollsys/tollmon/g"
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"fmt"
)

type Message struct {
	STX byte
	MC  []byte
	MT  []byte
	MB  []byte
	ETX byte
}

var (
	STX byte = 0x02 //Message Start
	ETX byte = 0x03 //Message End

	LenMC = 2
	LenMT = 2

	McData       = 0x01 //Data Message Catalog
	McAlert      = 0x20 //Alert Message Catalog
	McTest       = 0x30 //Test Message Catalog
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
	MtDutyEnd         = 0x03 //Offduty
	MtTypeChange      = 0x04 //Vehicle Type Changed
	MtEntryCard       = 0x05 //Entry Card Storge Min
	MtExitCard        = 0x06 //Exit Card Storge Max
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

	LenTime     = 14
	LenLaneID   = 26
	LenShift    = 2
	LenEmpID    = 4
	LenClass    = 2
	LenType     = 2
	LenPass     = 4
	LenLoan     = 4
	LenForfeit  = 4
	LenETCCar   = 1
	LenEmpName  = 20
	LenDutyTime = 20

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
	TIMEFOMAT = "2018-07-12 12:34:56.789"
)

func parseMsg(c chan byte) {
	for {
		var buffer []byte
		//g.LogDebug("parse msg start...")
		b := <-c
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
	}
}

func handleMsg(b []byte) {
	var msg Message
	index := 0
	msg.STX = b[index]
	index += 1
	msg.MC = b[index : index+LenMC]
	index += LenMC
	msg.MT = b[index : index+LenMT]
	index += LenMT
	msg.MB = b[index : len(b)-1]
	msg.ETX = b[len(b)-1]

	t:=hexStringToChs("CDF5D3F1D5E420202020")
	fmt.Println(t)
}

func HexStringToInt(s string) int {
	b, _ := hex.DecodeString(s)
	if len(b) == 1 {
		b = append([]byte{0x00, 0x00, 0x00}, b[0])
	}
	if len(b) == 2 {
		b = append([]byte{0x00, 0x00}, b[0], b[1])
	}
	if len(b) == 3 {
		b = append([]byte{0x00}, b[0], b[1],b[2])
	}
	if len(b) > 4{
		g.LogError("err length ",len(b))
	}
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int32
	binary.Read(bytesBuffer, binary.BigEndian, &tmp)
	return int(tmp)
}
func hexStringToChs(s string) string{
	gbkBs, _ := hex.DecodeString(s)
	i := bytes.NewBuffer(gbkBs)
	o := transform.NewReader(i,simplifiedchinese.GBK.NewDecoder())
	d,e := ioutil.ReadAll(o)
	if e != nil{
		g.LogError("hexStringToChs err",e.Error())
	}
	t := string(d)
	return t
}
//func HexStringToString(s string) string{
//	//b, _ := hex.DecodeString(s)
//}
