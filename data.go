package callposter

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	//CMCC 中国移动
	CMCC byte = iota + 0x01
	//CUCC 中国联通
	CUCC
	//CTCC 中国电信
	CTCC
	//CTCC_v 电信虚拟运营商
	CTCC_v
	//CUCC_v 联通虚拟运营商
	CUCC_v
	//CMCC_v 移动虚拟运营商
	CMCC_v
	intLength        = 4
	charLength       = 1
	headLength       = 8
	phoneIndexLength = 9
	phoneDataFile    = "phone.dat"
)

//PhoneRecord 号码属性信息
type PhoneRecord struct {
	PhoneNum string
	Province string
	City     string
	ZipCode  string
	AreaZone string
	CardType string
}

var (
	content []byte
	//CardTypemap 手机卡属性表
	CardTypemap = map[byte]string{
		CMCC:   "中国移动",
		CUCC:   "中国联通",
		CTCC:   "中国电信",
		CTCC_v: "中国电信虚拟运营商",
		CUCC_v: "中国联通虚拟运营商",
		CMCC_v: "中国移动虚拟运营商",
	}
	totalLen, firstOffset int32
)

func (pr PhoneRecord) String() string {
	return fmt.Sprintf("PhoneNum: %s\nAreaZone: %s\nCardType: %s\nCity: %s\nZipCode: %s\nProvince: %s\n", pr.PhoneNum, pr.AreaZone, pr.CardType, pr.City, pr.ZipCode, pr.Province)
}

func get4(b []byte) int32 {
	if len(b) < 4 {
		return 0
	}
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
}

func getN(s string) (uint32, error) {
	var n, cutoff, maxVal uint32
	i := 0
	base := 10
	cutoff = (1<<32-1)/10 + 1
	maxVal = 1<<uint(32) - 1
	for ; i < len(s); i++ {
		var v byte
		d := s[i]
		switch {
		case '0' <= d && d <= '9':
			v = d - '0'
		case 'a' <= d && d <= 'z':
			v = d - 'a' + 10
		case 'A' <= d && d <= 'Z':
			v = d - 'A' + 10
		default:
			return 0, errors.New("invalid syntax")
		}
		if v >= byte(base) {
			return 0, errors.New("invalid syntax")
		}

		if n >= cutoff {
			// n*base overflows
			n = (1<<32 - 1)
			return n, errors.New("value out of range")
		}
		n *= uint32(base)

		n1 := n + uint32(v)
		if n1 < n || n1 > maxVal {
			// n+v overflows
			n = (1<<32 - 1)
			return n, errors.New("value out of range")
		}
		n = n1
	}
	return n, nil
}

func totalRecord() int32 {
	return (int32(len(content)) - firstRecordOffset()) / phoneIndexLength
}

func firstRecordOffset() int32 {
	return get4(content[intLength : intLength*2])
}

//Find 二分法查询phone数据
func Find(phoneNum string) (pr *PhoneRecord, err error) {
	if len(phoneNum) < 7 || len(phoneNum) > 11 {
		return nil, errors.New("illegal phone length")
	}

	var left int32
	phoneSevenInt, err := getN(phoneNum[0:7])
	if err != nil {
		return nil, errors.New("illegal phone number")
	}
	phoneSevenInt32 := int32(phoneSevenInt)
	right := (totalLen - firstOffset) / phoneIndexLength
	for {
		if left > right {
			break
		}
		mid := (left + right) / 2
		offset := firstOffset + mid*phoneIndexLength
		if offset >= totalLen {
			break
		}
		curPhone := get4(content[offset : offset+intLength])
		recordOffset := get4(content[offset+intLength : offset+intLength*2])
		cardType := content[offset+intLength*2 : offset+intLength*2+charLength][0]
		switch {
		case curPhone > phoneSevenInt32:
			right = mid - 1
		case curPhone < phoneSevenInt32:
			left = mid + 1
		default:
			cbyte := content[recordOffset:]
			endOffset := int32(bytes.Index(cbyte, []byte("\000")))
			data := bytes.Split(cbyte[:endOffset], []byte("|"))
			cardStr, ok := CardTypemap[cardType]
			if !ok {
				cardStr = "未知电信运营商"
			}
			pr = &PhoneRecord{
				PhoneNum: phoneNum,
				Province: string(data[0]),
				City:     string(data[1]),
				ZipCode:  string(data[2]),
				AreaZone: string(data[3]),
				CardType: cardStr,
			}
			return
		}
	}
	return nil, errors.New("phone's data not found")
}
