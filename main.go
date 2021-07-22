package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	mixin "github.com/fox-one/mixin-sdk-go"
	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/montanaflynn/stats"
	"github.com/tidwall/gjson"
)

//mixin bot config
const (
	ClientID   = ""
	SessionID  = ""
	PrivateKey = ""
	PinToken   = ""
	Pin        = ""
)

//coingecko api host
const apihost = "https://api.coingecko.com/api/v3"

//Database
func createdb() {
	log.Println("Creating sqlite db...")
	file, err := os.Create("sqlite.db")
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
	log.Println("sqlite db created")
}

func createTable(db *sql.DB) {
	createTablesql := `CREATE TABLE subuser(
		"UserID" TEXT NOT NULL UNIQUE,
		"ConversationID" TEXT,
		"Sub" integer
	);`
	statement, err := db.Prepare(createTablesql)
	if err != nil {
		log.Fatalln(err)
	}
	statement.Exec()
	log.Println("subuser table created")
}

func checkSubUser(db *sql.DB, UserID string) bool {
	var inStatus bool
	checkSubusersql := fmt.Sprintf("SELECT UserID FROM subuser WHERE UserID = '%s'", UserID)
	row, err := db.Query(checkSubusersql)
	if err != nil {
		log.Fatalln(err)
	}
	defer row.Close()
	var userid string
	for row.Next() {
		row.Scan(&userid)
		if len(userid) != 0 {
			inStatus = true
		} else if len(userid) == 0 {
			inStatus = false
		}
	}
	return inStatus
}

func insertSubuser(db *sql.DB, UserID string, ConversationID string, Sub bool) {
	insertSubuersql := `INSERT OR IGNORE INTO subuser(UserID, ConversationID, Sub) VALUES (?,?,?)`
	statement, err := db.Prepare(insertSubuersql)
	if err != nil {
		log.Fatalln("1", err)
	}
	_, err = statement.Exec(UserID, ConversationID, Sub)
	if err != nil {
		log.Fatalln("2", err)
	}
}

func controlSub(db *sql.DB, toggle, UserID string) {
	updateSubsql := fmt.Sprintf("UPDATE subuser SET Sub = '%s' WHERE UserID = '%s'", toggle, UserID)
	statement, err := db.Prepare(updateSubsql)
	if err != nil {
		log.Fatalln("1", err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatalln("2", err)
	}
	log.Println(toggle + " sub succeed")
}

func deleteSubuser(db *sql.DB, UserID string) {
	deleteSubusersql := `DELETE FROM subuser WHERE UserID = ?`
	statement, err := db.Prepare(deleteSubusersql)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = statement.Exec(UserID)
	if err != nil {
		log.Fatalln(err)
	}
}

func displaySubuser(db *sql.DB) {
	row, err := db.Query("SELECT * FROM subuser ORDER BY Sub")
	if err != nil {
		log.Fatalln(err)
	}
	defer row.Close()
	for row.Next() {
		var userid string
		var convertionid string
		var sub bool
		row.Scan(&userid, &convertionid, &sub)
		fmt.Println("userid:", userid, "\nconversation:", convertionid, "\nsub:", sub)
	}
}

func statusSubuser(db *sql.DB, UserID string) string {
	var returnstring string
	statusSubusersql := fmt.Sprintf("SELECT Sub FROM subuser WHERE UserID = '%s'", UserID)
	rows, err := db.Query(statusSubusersql)
	if err != nil {
		log.Fatalln("1", err)
	}
	defer rows.Close()
	var status bool
	for rows.Next() {
		if err := rows.Scan(&status); err == nil {
			if status {
				returnstring = "订阅状态: 已订阅"
			} else if !status {
				returnstring = "订阅状态: 未订阅"
			}
		} else {
			log.Println("err:", err)
		}
	}
	return returnstring
}

func CoingeckoMarketChartRange(id, vs_currency, from, to string) string {
	api := apihost + "/coins/" + id + "/market_chart/range" + "?id=" + id + "&vs_currency=" + vs_currency + "&from=" + from + "&to=" + to
	resp, err := http.Get(api)
	if err != nil {
		log.Println(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	data := string(body)
	return data
}

//Coingecko
func CoingeckoHistory(id string, date time.Time) string {
	t := fmt.Sprintf("%d-%02d-%4d", date.Day()-1, date.Month(), date.Year())
	api := apihost + "/coins/" + id + "/history" + "?date=" + t
	resp, err := http.Get(api)
	if err != nil {
		log.Println(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	data := string(body)
	return data
}

func CoingeckoPrice(id, vs_currencies string) string {
	api := apihost + "/simple/price" + "?ids=" + id + "&vs_currencies=" + vs_currencies
	resp, err := http.Get(api)
	if err != nil {
		log.Println(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	data := string(body)
	return data
}

func Mixinrespond(client *mixin.Client, ctx context.Context, msg *mixin.MessageView, category string, data []byte, step int) error {
	payload := base64.StdEncoding.EncodeToString(data)
	id, _ := uuid.FromString(msg.MessageID)
	reply := &mixin.MessageRequest{
		ConversationID: msg.ConversationID,
		RecipientID:    msg.UserID,
		MessageID:      uuid.NewV5(id, fmt.Sprintf("reply %d", step)).String(),
		Category:       category,
		Data:           payload,
	}
	log.Println("respond: ", string(data), "->", msg.UserID)
	return client.SendMessage(ctx, reply)
}

func MixinMsg(client *mixin.Client, ctx context.Context, data []byte, ConversationID, RecipientID string) error {
	payload := base64.StdEncoding.EncodeToString(data)
	messageUuid, _ := uuid.NewV4()
	reply := &mixin.MessageRequest{
		ConversationID: ConversationID,
		RecipientID:    RecipientID,
		MessageID:      messageUuid.String(),
		Category:       "PLAIN_TEXT",
		Data:           payload,
	}
	return client.SendMessage(ctx, reply)
}

func MixinSubBroadcast(db *sql.DB, client *mixin.Client, ctx context.Context, data []byte) {
	row, err := db.Query("SELECT * FROM subuser ORDER BY Sub")
	if err != nil {
		log.Fatalln(err)
	}
	defer row.Close()
	for row.Next() {
		var userid string
		var convertionid string
		var sub bool
		row.Scan(&userid, &convertionid, &sub)
		if sub {
			MixinMsg(client, ctx, data, convertionid, userid)
		}
	}
}

func getahr999() float64 {
	now := time.Now()
	nowux := now.Unix()
	before := nowux - 24*200*60*60
	nowuxs := strconv.Itoa(int(nowux))
	befores := strconv.Itoa(int(before))
	valueslice := []float64{}
	data := CoingeckoMarketChartRange("bitcoin", "usd", befores, nowuxs)
	values := gjson.Get(data, "prices.#.1").Array() //200 day data list
	for _, xd := range values {
		valueslice = append(valueslice, xd.Num)
	}
	avg, err := stats.HarmonicMean(valueslice)
	if err != nil {
		log.Fatal(err)
	}
	js1 := CoingeckoPrice("bitcoin", "usd")
	price := gjson.Get(js1, "bitcoin.usd").Float()
	bornday := (nowux - 1230940800) / (24 * 60 * 60)
	logprice := math.Pow(10, 5.84*math.Log10(float64(bornday))-17.01)
	ahr999 := math.Round((price/avg)*(price/logprice)*1000) / 1000
	return ahr999
}

func getahr999string() string {
	now := time.Now()
	nowux := now.Unix()
	before := nowux - 24*200*60*60
	nowuxs := strconv.Itoa(int(nowux))
	befores := strconv.Itoa(int(before))
	valueslice := []float64{}
	data := CoingeckoMarketChartRange("bitcoin", "usd", befores, nowuxs)
	values := gjson.Get(data, "prices.#.1").Array() //200 day data list
	for _, xd := range values {
		valueslice = append(valueslice, xd.Num)
	}
	avg, err := stats.HarmonicMean(valueslice)
	if err != nil {
		log.Fatal(err)
	}
	avgs := fmt.Sprintf("%.3f", avg)
	js1 := CoingeckoPrice("bitcoin", "usd")
	price := gjson.Get(js1, "bitcoin.usd").Float()
	prices := gjson.Get(js1, "bitcoin.usd").String()
	bornday := (nowux - 1230940800) / (24 * 60 * 60)
	logprice := math.Pow(10, 5.84*math.Log10(float64(bornday))-17.01)
	logprices := fmt.Sprintf("%.3f", logprice)
	ahr999 := math.Round((price/avg)*(price/logprice)*1000) / 1000
	ahr999s := fmt.Sprintf("%.3f", ahr999)
	var section string
	if ahr999 <= 0.45 {
		section = "当前区间: 抄底区间"
	} else if ahr999 > 0.45 && ahr999 <= 1.2 {
		section = "当前区间: 定投区间"
	} else if ahr999 > 1.2 && ahr999 <= 5 {
		section = "当前区间: 坐稳起飞区间"
	} else if ahr999 > 5 {
		section = "当前区间：已起飞区间"
	}
	datastring := "当前价格:" + prices + "\n200日定投成本:" + avgs + "\n拟合价格:" + logprices + "\nAhr999指数:" + ahr999s + "\n" + section
	return datastring
}

func message() {
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite.db")
	defer sqliteDatabase.Close()
	ctx := context.Background()
	s := &mixin.Keystore{
		ClientID:   ClientID,
		SessionID:  SessionID,
		PrivateKey: PrivateKey,
		PinToken:   PinToken,
	}
	h := func(ctx context.Context, msg *mixin.MessageView, userID string) error {
		client, err := mixin.NewFromKeystore(s)
		if err != nil {
			log.Fatal(err)
		}
		if userID, _ := uuid.FromString(msg.UserID); userID == uuid.Nil {
			return nil
		}
		data, err := base64.StdEncoding.DecodeString(msg.Data)
		if err != nil {
			return err
		}
		log.Println("Message:", string(data))

		introbutton := `{"label": "使用介绍", "action": "input:?", "color": "#5979F0"}`
		subbutton := `{"label": "点我订阅", "action": "input:/sub", "color": "#5979F0"}`
		unsubbutton := `{"label": "点我退订", "action": "input:/unsub", "color": "#B76753"}`
		statusbutton := `{"label": "订阅状态", "action": "input:/status", "color": "#6BC0CE"}`
		ahr999introbutton := `{"label": "Ahr999指数介绍", "action": "input:/ahr999intro", "color": "#75A2CB"}`
		helpmessagePost := `
## 机器人介绍:
Ahr999指数订阅机器人，订阅后每24小时播报一次ahr999指数，点击机器人按钮可以查看指数的历史图表。

## 命令:
- ahr999	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;获取现在的ahr999指数
- ahr999intro	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;ahr999指数的介绍文章
- /sub	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;订阅ahr999指数播报(每30分钟播报一次)
- /unsub	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;取消播报 
- ?		&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;命令列表
- 点击机器人图标    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;ahr999指数图表
`
		ahr999introPost := `
## Ahr999 介绍    

<img src="https://ahr999.com/ahr999.jpg" width = "200" height = "50%" alt="ahr999头像" align=center />

该文集为九神[@ahr999](https://weibo.com/ahr999)在过去几年囤比特币过程中的思考和经验。写它的目的不是宣传比特币，而是帮助那些已经准备囤比特币的人。是否看好比特币很大程度上与价值观有关，无意改变任何人的价值观。（原整理[@玛雅cndx](http://jdoge.com/)）
  

# 新文：

- 《[囤比特币：ahr999指数](https://weibo.com/ttarticle/p/show?id=2309404441088189399138)》。可以用ahr999指数机器人查询ahr999指数。
- 《[寻找合适的购买时机（20190804更新）](https://weibo.com/ttarticle/p/show?id=2309404401520245342246)》囤币党也是要抄底的。对（中长期）行情精准的判断，也是囤币党的必备技能。

# 序章：

- 《[知之非难，行之不易](https://weibo.com/ttarticle/p/show?id=2309404290257041409981)》为什么要写这个系列。

# 入门：

- 《[比特币与理想主义](https://weibo.com/ttarticle/p/show?id=2309404283412763544904)》我们在参与一场社会实验，它存在失败的可能性，但是我们无怨无悔。
- 《[下车太早只因愿景太小](https://weibo.com/ttarticle/p/show?id=2309404286329633561927)》这场实验的目标很大，如果一切顺利，比特币的价格可能会在20年后涨到1.6亿人民币。
- 《[囤比特币：你离财富自由还有多远？](https://weibo.com/ttarticle/p/show?id=2309404287022729712573)》我们没有其他能耐，只能靠囤积比特币，并耐心地等待属于自己的财富自由。
- 《[囤比特币：冲动、孤独、无聊与矛盾](https://weibo.com/ttarticle/p/show?id=2309404287827880877926)》虽然会经历冲动、孤独、无聊和矛盾等心理考验，但是我们已经做好准备囤币。
- 《[囤比特币：手握私钥的快感](https://weibo.com/ttarticle/p/show?id=2309404289198575222102)》虽然掌握私钥有点麻烦，但是我们仍然准备自己对自己负责。
- 《[囤比特币：如何管理私钥？](https://weibo.com/ttarticle/p/show?id=2309404289950832033282)》管理私钥其实并没有想象的那么麻烦，但我们需要把握好几个原则。

# 进阶：

- 《[囤比特币：基本价格模型](https://weibo.com/ttarticle/p/show?id=2309404290588110395875)》囤币是比特币一切价值得来源，长期囤币者关心的去中心化和安全性是比特币首先要保证的特性。每次产量减半，囤币需求不变，但比特币供应减少，价格必涨。
- 《[囤比特币：寻找合适的购买时机](https://weibo.com/ttarticle/p/show?id=2309404292613674022595)》我们都希望使用有限的投入获得更多比特币，那么何时是合适的买点呢？【[Ahr999图表](http://ahr999mixin.tk)】
- 《[囤比特币：唯有比特币](https://weibo.com/ttarticle/p/show?id=2309404294325361104197)》除了比特币，我不持有任何其它数字币。但是，我也不反对任何人持有任何币，哪怕是传销币。
- 《[囤比特币：不要跟着感觉走](https://weibo.com/ttarticle/p/show?id=2309404294599689565825)》囤比特币其实是反复决策的结果，别人觉得简单是因为只看到结果，而看不到决策的过程。
- 《[囤比特币：币本位思维](https://weibo.com/ttarticle/p/show?id=2309404294635697610801)》比特币创造了一个全新的世界。在这个世界里，只有一个标准——比特币。【[币本位USD/XBT](http://btcie.com/btc)】
- 《[囤比特币：心中无币](https://weibo.com/ttarticle/p/show?id=2309404295149122413875)》当理解上升到一定程度，我们不再需要关注任何比特币相关的信息。

# 贡献：

- 《[囤比特币：打造强节点](https://weibo.com/ttarticle/p/show?id=2309404297578786198023)》成就最好的自己就是对比特币最大的贡献！
- 《[囤比特币：运行全节点](https://weibo.com/ttarticle/p/show?id=2309404297617780650574)》私钥决定比特币所有权，全节点捍卫比特币规则。

# 终章：

- 《[不忘初心](https://weibo.com/ttarticle/p/show?id=2309404297653562298410)》系列的最后的一文。

# 故事：

- 《[四年一个轮回，不光有世界杯，还有比特币](https://weibo.com/ttarticle/p/show?id=2309404265822628505977)》有时候，时间能改变一切。有时候，时间什么也改变不了。
- 《[上一轮熊市](https://weibo.com/ttarticle/p/show?id=2309404282406097046246)》从8000元到900元，我们都经历了些什么？
- 《[牛市起点的故事](https://weibo.com/ttarticle/p/show?id=2309404284267738876518)》2015年11月，牛市起点，19个人，19个故事。
`

		switch string(data) {
		// for dev
		case "showid":
			covidString := msg.ConversationID
			useridString := msg.UserID
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte(covidString), 2)
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte(useridString), 3)
		case "/display":
			displaySubuser(sqliteDatabase)

		// usages
		case "ahr999":
			ahr := getahr999string()
			log.Println(ahr)
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte(ahr), 4)
		case "/sub":
			s := checkSubUser(sqliteDatabase, msg.UserID)
			if s {
				controlSub(sqliteDatabase, "true", msg.UserID)
			} else if !s {
				insertSubuser(sqliteDatabase, msg.UserID, msg.ConversationID, true)
			}
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte("订阅成功！您将会收到机器人广播的新消息。"), 5)
			displaySubuser(sqliteDatabase)
		case "/unsub":
			controlSub(sqliteDatabase, "false", msg.UserID)
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte("取消订阅成功！您将不会收到来自机器人的消息。（您还可以用/del删除您在数据库中的用户记录。)"), 6)
			displaySubuser(sqliteDatabase)
		case "/del":
			deleteSubuser(sqliteDatabase, msg.UserID)
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte("删除记录成功！"), 6)
			displaySubuser(sqliteDatabase)
		case "/status":
			status := statusSubuser(sqliteDatabase, msg.UserID)
			log.Println(status)
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte(status), 6)
			displaySubuser(sqliteDatabase)

		//help
		case "?", "/?", "？", "/？":
			msgbutton := "[" + subbutton + "," + unsubbutton + "," + statusbutton + "," + ahr999introbutton + "]"
			Mixinrespond(client, ctx, msg, mixin.MessageCategoryPlainPost, []byte(helpmessagePost), 7)
			if err := Mixinrespond(client, ctx, msg, mixin.MessageCategoryAppButtonGroup, []byte(msgbutton), 8); err != nil {
				log.Println(err)
			}
		case "Hi", "你好":
			msgbutton := "[" + introbutton + "," + subbutton + "," + unsubbutton + "," + statusbutton + "," + ahr999introbutton + "]"
			if err := Mixinrespond(client, ctx, msg, mixin.MessageCategoryAppButtonGroup, []byte(msgbutton), 8); err != nil {
				log.Println(err)
			}
		case "/ahr999intro":
			if err := Mixinrespond(client, ctx, msg, mixin.MessageCategoryPlainPost, []byte(ahr999introPost), 8); err != nil {
				log.Println(err)
			}
		default:
			ahr := getahr999string()
			msgbutton := "[" + introbutton + "," + subbutton + "," + unsubbutton + "," + statusbutton + "," + ahr999introbutton + "]"
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte(ahr), 4)
			if err := Mixinrespond(client, ctx, msg, mixin.MessageCategoryAppButtonGroup, []byte(msgbutton), 8); err != nil {
				log.Println(err)
			}
		}
		return nil
	}

	client, err := mixin.NewFromKeystore(s)
	if err != nil {
		log.Fatal(err)
	}
	for {
		if err := client.LoopBlaze(ctx, mixin.BlazeListenFunc(h)); err != nil {
			log.Printf("LoopBlaze: %v", err)
		}
		time.Sleep(time.Second)
	}
}

func main() {
	// message module
	go message()
	// check if database file exist
	if _, err := os.Stat("sqlite.db"); os.IsNotExist(err) {
		createdb()
		sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite.db")
		defer sqliteDatabase.Close()
		createTable(sqliteDatabase)
	}
	sqliteDatabase, _ := sql.Open("sqlite3", "./sqlite.db")
	defer sqliteDatabase.Close()
	ctx := context.Background()
	s := &mixin.Keystore{
		ClientID:   ClientID,
		SessionID:  SessionID,
		PrivateKey: PrivateKey,
		PinToken:   PinToken,
	}
	client, err := mixin.NewFromKeystore(s)
	if err != nil {
		log.Fatal(err)
	}

	// main loop
	for {
		num := getahr999()
		index := getahr999string()
		log.Println(index)
		if num <= 0.45 {
			index := "当前指数已达抄底线!\n" + index
			MixinSubBroadcast(sqliteDatabase, client, ctx, []byte(index))
		}
		MixinSubBroadcast(sqliteDatabase,client, ctx, []byte(index))
		time.Sleep(time.Hour * 24)
	}
}
