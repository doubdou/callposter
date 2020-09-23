package callposter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

//主叫外显
//根据配置文件"外呼主叫号码表.xlsx"配置轮询
//注意：主叫号码表限制最大并发数

//外呼前缀
//一、归属地规则：查询号码归属地，外地号码加0，本地号码不加前缀
//二、落地网关规则：自定义前缀，用于路由选择，比如被叫号码是15606130692，网关选路使用aidev前缀，则送呼设置被叫为aidev15606130692
//其他规则暂无定义

var mng = struct {
	*sync.Mutex
	mapping map[string]bool
}{
	Mutex:   new(sync.Mutex),
	mapping: make(map[string]bool),
}

var callerDisplaySum int

type putRequestMsg struct {
	CallerDisplay string `json:"caller_display"` //主叫外显
}

type getResponseMsg struct {
	CallerDisplay string `json:"caller_display"` //主叫外显
	ActualCallee  string `json:"actual_callee"`  //外呼的被叫号码
}

//释放接口
// 在map中设置未使用 false
func doPut(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	if len(buf) == 0 {
		w.WriteHeader(400)
		return
	}
	req := putRequestMsg{}
	err = json.Unmarshal(buf, &req)
	mng.Lock()
	if mng.mapping[req.CallerDisplay] == false {
		log.Println("do put callerDisplay:", req.CallerDisplay, "not used!")
		w.WriteHeader(400)
	} else {
		mng.mapping[req.CallerDisplay] = false
		log.Println("do put callerDisplay:", req.CallerDisplay, "using flag: true -> false")
	}
	mng.Unlock()
}

//查询接口
func doGet(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	cnt := 0
	var resp getResponseMsg

	if len(vars["callee"]) == 0 {
		w.WriteHeader(400)
		log.Println("get fail: callee not found in url!")
		return
	}
	callee := vars["callee"][0]
	mng.Lock()
	for k, v := range mng.mapping {
		if false == v {
			//设置使用中
			mng.mapping[k] = true
			resp.CallerDisplay = k
			log.Println("do Get callerDisplay:", k, "using flag: false -> true")
			break
		}
		cnt++
		if cnt >= callerDisplaySum {
			break
		}
	}
	mng.Unlock()

	//无可用主叫外显
	if cnt >= callerDisplaySum {
		w.WriteHeader(400)
		log.Println("No caller Display available!")
		return
	}
	pr, err := Find(callee)
	//查询归属地信息失败
	if err != nil {
		w.WriteHeader(400)
		log.Println("Search callee attribution fail!")
		return
	}
	if pr.City == attributionConfigGet().Location {
		//命中
		resp.ActualCallee = attributionConfigGet().HitPreifx + pr.PhoneNum
	} else {
		//未命中
		resp.ActualCallee = attributionConfigGet().MissPrefix + pr.PhoneNum
	}
	fmt.Println(pr)
	jsonStr, _ := json.Marshal(&resp)
	w.Header().Set("content-type", "application/json")
	fmt.Fprintf(w, string(jsonStr))
}

func doEntry(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		doPut(w, r)
	} else if r.Method == "GET" {
		doGet(w, r)
	}
}

// ListenAndServe 服务启动入口
func ListenAndServe() {
	//启动前文件处理
	err := doFile()
	if err != nil {
		log.Println("http serve start fail. exit...")
		return
	}
	//服务启动
	addr := fmt.Sprintf("%s:%s", httpConfigGet().IP, httpConfigGet().Port)

	http.HandleFunc("/v1/voip/callposter/outgoing", doEntry)
	log.Println("callposter serve at ", addr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Printf("ListenAndServe() exexute failed, %v\n", err)
		return
	}
}
