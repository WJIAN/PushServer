package main

import (
	"time"
	"strings"
	"fmt"
	"bytes"
    "net/http"
	"io/ioutil"
	"os"

	"PushServer/slog"
	"PushServer/rediscluster"
)


type sendState struct {
	// push 计数
	count int64
	// 上一次推送rest的返回内容
	rvLast string

}


var redisPool *rediscluster.RedisPool



func getClients() []string {
	fun := "updateMan"
    cmd := []interface{}{
		"keys",
		"I.*",
	}


	mcmd := make(map[string][]interface{})
    mcmd["10.241.221.106:9600"] = cmd

	rp := redisPool.Cmd(mcmd)
	slog.Debugln(fun, "evalsha1", rp)


	clientlist := []string{}
	for addr, r := range(rp) {
		cl, err := r.List()
		if err != nil {
			slog.Errorf("%s addr:%s err:%s", fun, addr, err)
			break
		}

		for _, v := range(cl) {
			clientlist = append(clientlist, strings.Split(v, ".")[1])
		}


		break
	}


	return clientlist

}

// 业务数据包发送
func restPush(clientid string, sendData []byte) string {
	fun := "restPush"

    //sendTime := time.Now().Format("2006-01-02 15:04:05")
	//sendData := []byte(sendTime+"您您您您您您您您")

	slog.Infof("%s cid:%s len:%d data:%s", fun, clientid, len(sendData), sendData)
	
	client := &http.Client{}
	url := fmt.Sprintf("http://42.120.4.112:9090/push/%s/0/0", clientid)
	reqest, _ := http.NewRequest("POST", url, bytes.NewReader(sendData))

	reqest.Header.Set("Connection","Keep-Alive")

	response,_ := client.Do(reqest)
	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			slog.Errorf("%s cid:%s Push return ERROR %s", fun, clientid, err)
			return fmt.Sprintf("push err:%s", err)
		}

		slog.Infof("%s cid:%s Push return %s", fun, clientid, body)
		return string(body)

	} else {
		slog.Errorf("%s cid:%s Push ERROR", fun, clientid)
		return fmt.Sprintf("push errcode:%s", response.StatusCode)

	}


}

func updateMan(sendMan map[string]*sendState) {
	clis := getClients()
	slog.Infoln(clis)
	for _, c := range(clis) {
		if _, ok := sendMan[c]; !ok {
			sendMan[c] = &sendState {
				count: 0,
				rvLast: "",
			}

		}
	}

	slog.Infof("%s sendMan len:%d", "updateMan", len(sendMan))

}

func traversePush(sendMan map[string]*sendState) {
	joke := `In Go servers, each incoming request is handled in its own goroutine. Request handlers often start additional goroutines to access backends such as databases and RPC services. The set of goroutines working on a request typically needs access to request-specific values such as the identity of the end user, authorization tokens, and the request's deadline. When a request is canceled or times out, all the goroutines working on that request should exit quickly so the system can reclaim any resources they are using.`

	joke = "你好"

	myjoke1 := `
一晚，无名师和Nubi参加一个程序员的探讨会。有个程序员问Nubi和他的老师来自哪个学校。当得知他们是Unix大道的追随者时候，程序员颇为不屑。 “Unix命令行工具太粗糙，太落后”，他讥讽道。“现代的、设计得当的操作系统可以在图形用户界面中做任何事情。” 无名师一言不发，只是指着月亮。旁边的一条狗对着他的手狂吠。 “我不明白。”程序员说。无名师依然缄默，指着一副佛祖像，然后又指着一扇窗。 “你想说什么”程序员问。无名师指着程序员的头，接着指着一块大石。 “请把话说清楚！”程序员要求到。无名师深深蹙眉，轻拍程序员的鼻子两下，把他扔到旁边的垃圾箱中。程序员试图从垃圾堆挣扎出来时，那条狗过来在他身上便溺。此时，程序员眼中一亮！
`

	myjoke2 := `
一九四五年的一天，克力富兰的孤儿院里出现了一个神秘的女婴，没有人知道她的父母是谁。她孤独地长大，没有任何人与她来往。 
直到一九六三年的一天，她莫明其妙地爱上了一个流浪汉，情况才变得好起来。可是好景不长，不幸事件一个接一个的发生。 
首先，当她发现自己怀上了流浪汉的小孩时，流浪汉却突然失踪了。其次，她在医院生小孩时，医生发现她是双性人，也就是说她同时具有男女性器官。为了挽救她的生命，医院给她做了变性手术，她变成了他。最不幸的是，她刚刚生下的小女孩又被一个神秘的人给绑走了。这一连串的打击使他从此一蹶不振，最后流落到街头变成了一个无家可归的流浪汉。 
直到一九七八年的一天，他醉熏熏地走进了一个小酒吧，把他一身不幸的遭遇告诉了一个比他年长的酒吧伙计。酒吧伙计很同情他，主动提出帮他找到那个使 他 怀孕而又失踪的流浪汉。唯一的条件是他必须参加伙计他们的时间旅行特种部队 。他们一起进了 时间飞车 。飞车回到六三年时，伙计把流浪汉放了出去。流浪汉莫明其妙地爱上了一个孤儿院长大的姑娘，并使她怀了孕。伙计又乘 时间飞车 前行九个多月，到医院抢走了刚刚出生的小女婴，并用 时间飞车 把女婴带回到一九四五年，悄悄地把她放在克力富兰的一个孤儿院里。然后再把稀里糊涂的流浪汉向前带到了一九八五年，并且让他加入了他们的 时间旅行特种部队 。 
流浪汉有了正式工作以后，生活走上了正轨。并逐渐地在特种部队里混到了相当不错的地位。有一次，为了完成一个特殊任务，上级派他飞回一九七零年，化装成酒吧伙计去拉一个流浪汉加入他们的特种部队.
`

	myjoke3 := `
我们可能永远不会知道IBM在1979年选择Intel的8088（一种与8086同代的8位芯片）作为它新开发的PC的CPU的原因。从技术上说，当时有很多公司可以提供更出色的方案，如Motorala和National Semiconductor。由于选择了Intel的芯片，IBM帮助Intel在接下来的20多年里面财源滚滚，就像IBM选择了Microsoft的MS-DOS作为PC的操作系统从而使Microsoft飞黄腾达一样。具有讽刺意味的是，1993年8月，Intel的股票市值达到了266亿美元，超过了IBM的245亿美元，从而取代了IBM成为美国市值最高的电子类公司。

Intel和Microsoft凭借其独家的经营产品，获得了远远超出其贡献的暴利！成为新的“IBM”。而IBM还在绝望中挣扎，试图恢复自己以前的地位。他推出了PowerPC，企图打破Intel在硬件的垄断，同时推出OS/2操作系统，试图动摇Microsof在软件上的统治。OS/2失败无疑，但断定PowerPC的厄运则为时尚早。

……

IBM在PC上所做的部分决定（也许是大部分决定）显然是出自非技术背景，在决定采用MS-DOS之前，IBM安排了一个会议，与Digital Research公司的Gray Kildall 商讨CP/M操作系统的事宜。就在会议举行的当天，出现了人们传说的故事：由于天气非常好，Gray Kildall决定改坐自己的直升飞机与会，结果误点。IBM的经理们肯能对长时间等待颇感恼火，便转而与Microsoft匆匆达成了协议。

Bill Gates当时刚从Seattle Computer Product公司购买了QDOS（从文学效果上说它表示Quick and Dirty Operating System 快速而肮脏的操作系统），他对它稍作整理后，更名为MS-DOS。接下来的故事，都已是人们“津津乐道”的历史典故。IBM很高兴，Intel也很高兴Microsoft则非常非常高兴。Digital Research自然不会愉快。数年以后，Seattle Computer Product意识到了自己放走了一个有史以来销量最大的计算机程序之后，自然也不会愉快。他们仍然保留了一个权利，就是在他们在销售硬件时可以同时销售MS-DOS。这就是为什么过去你看到一些出自Seattle Computer Product的MS-DOS的原因，他们被滑稽地附在显示已经没有用的Intel芯片产品上，从而“庄严”地履行他们与Microsoft达成的协议。

不要为Seattle Computer Product感到太遗憾——他们的QDOS本身很大程度上也是基于Gray Kildall的CP/M。而Gray Kildall似乎更热衷于飞行。Bill Gates后来用销售软件的利润购买了一辆性能出众，快如闪电的保时捷959跑车，花了75万美元，但在进入美国海关时却出了问题。保时捷959无法在美国驾驶，因为它没有通过美国政府规定必须进行的防撞性能测试。这辆跑车至今仍然搁置在奥克兰的一个仓库里，从来没有驾驶过，这大概是Bill Gates唯一可以保证不会崩溃的产品。

……

                                                ——摘自《C专家编程》

                                                  Expert C Programming

                                                Peter Van Der Linden
`

	fmt.Println(len(joke), len(myjoke1), len(myjoke2), len(myjoke3))
	for c, v := range(sendMan) {
		sendTime := time.Now().Format("01-02 15:04:05")
		var sendData string
		if c == "301239afbfdfcae2266337b7e2caaaa3" {
			sendData = fmt.Sprintf("S:%d/D:%s/LR:%s/M:[%s]", v.count, sendTime, v.rvLast, myjoke3)
		} else {
			sendData = fmt.Sprintf("S:%d/D:%s/LR:%s/M:[%s]", v.count, sendTime, v.rvLast, joke)
		}
		v.count++
		v.rvLast = restPush(c, []byte(sendData))
	}




}

func loopPush() {
	sendMan := make(map[string]*sendState)
	ticker := time.NewTicker(time.Second * 60)
	ticker2 := time.NewTicker(time.Second * 60 * 5)


	// 每5分更新一次client id
	updateMan(sendMan)
	go func() {
		for {
			select {
			case <-ticker2.C:
				updateMan(sendMan)
			}
		}
    }()


	// 循环遍历推送
	traversePush(sendMan)
	for {
		select {
		case <-ticker.C:
			traversePush(sendMan)
		}
	}

}



func main() {
    slog.Init(os.Stdout)
	redisPool = rediscluster.NewRedisPool()


	

	//restPush()

	loopPush()
}
