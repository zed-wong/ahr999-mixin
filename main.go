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
	"sync"
	"time"

	mixin "github.com/fox-one/mixin-sdk-go"
        "github.com/robfig/cron/v3"
	"github.com/gofrs/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/montanaflynn/stats"
	"github.com/tidwall/gjson"
)

//mixin bot config
const (
	ClientID   = 
	SessionID  = 
	PrivateKey = 
	PinToken   = 
	Pin        = 
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

func displaySubuser(db *sql.DB) string {
	/*
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
	           log.Println("userid:", userid, "\nconversation:", convertionid, "\nsub:", sub)
	   }
	*/
	length, err := db.Query("SELECT COUNT(UserID) FROM subuser")
	if err != nil {
		log.Fatalln(err)
	}
	defer length.Close()
	var lg string
	for length.Next() {
		length.Scan(&lg)
	}
	return lg
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
				returnstring = "????????????: ?????????"
			} else if !status {
				returnstring = "????????????: ?????????"
			}
		} else {
			log.Println("err:", err)
		}
	}
	return returnstring
}

//Coingecko
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

func MixinToMe(client *mixin.Client, ctx context.Context, data []byte) error {
	payload := base64.StdEncoding.EncodeToString(data)
	messageUuid, _ := uuid.NewV4()
	reply := &mixin.MessageRequest{
		ConversationID: "8169cfc6-a6f1-37bf-8ad0-d3b3ea99a5e5", //7000103262
		RecipientID:    "44d9717d-8cae-4004-98a1-f9ad544dcfb1", //28865
		MessageID:      messageUuid.String(),
		Category:       "PLAIN_TEXT",
		Data:           payload,
	}
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

func goMixinMsg(client *mixin.Client, ctx context.Context, data []byte, ConversationID, RecipientID string, wg *sync.WaitGroup) error {
	payload := base64.StdEncoding.EncodeToString(data)
	messageUuid, _ := uuid.NewV4()
	reply := &mixin.MessageRequest{
		ConversationID: ConversationID,
		RecipientID:    RecipientID,
		MessageID:      messageUuid.String(),
		Category:       "PLAIN_TEXT",
		Data:           payload,
	}
	defer wg.Done()
	return client.SendMessage(ctx, reply)
}

func MixinSubBroadcast(db *sql.DB, client *mixin.Client, ctx context.Context, data []byte) {
	row, err := db.Query("SELECT * FROM subuser ORDER BY Sub")
	if err != nil {
		log.Fatalln(err)
	}
	defer row.Close()
	length, err := db.Query(`select count(*) from subuser where sub = "1" or sub = "true"`)
	if err != nil {
		log.Fatalln(err)
	}
	defer length.Close()
	lg := checkCount(length)
	log.Println("Peoples:", lg)
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

func goMixinSubBroadcast(db *sql.DB, client *mixin.Client, ctx context.Context, data []byte, wg *sync.WaitGroup) {
	row, err := db.Query("SELECT * FROM subuser ORDER BY Sub")
	if err != nil {
		log.Fatalln(err)
	}
	defer row.Close()
	length, err := db.Query(`select count(*) from subuser where sub = "1" or sub = "true"`)
	if err != nil {
		log.Fatalln(err)
	}
	defer length.Close()
	lg := checkCount(length)
	wg.Add(lg)
	for row.Next() {
		var userid string
		var convertionid string
		var sub bool
		row.Scan(&userid, &convertionid, &sub)
		if sub {
			go goMixinMsg(client, ctx, data, convertionid, userid, wg)
		}
	}
	wg.Wait()
}

func checkCount(rows *sql.Rows) (count int) {
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			log.Println(err)
		}
	}
	return count
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

func getahr999x() float64 {
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
	ahr999x := math.Round(((avg/price)*(logprice/price)*3)*1000) / 1000
	return ahr999x
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
	avg, err := stats.GeometricMean(valueslice)
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
		section = "????????????: ????????????"
	} else if ahr999 > 0.45 && ahr999 <= 1.2 {
		section = "????????????: ????????????"
	} else if ahr999 > 1.2 && ahr999 <= 5 {
		section = "????????????: ??????????????????"
	} else if ahr999 > 5 {
		section = "??????????????????????????????"
	}
	datastring := "????????????:" + prices + "\n200???????????????:" + avgs + "\n????????????:" + logprices + "\nAhr999??????:" + ahr999s + "\n" + section
	return datastring
}

func getahr999xstring() string {
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
	ahr999x := math.Round(((avg/price)*(logprice/price)*3)*1000) / 1000
	ahr999s := fmt.Sprintf("%.3f", ahr999x)
	var section string
	if ahr999x <= 0.45 {
		section = "????????????: ????????????"
	} else if ahr999x > 0.45 && ahr999x <= 1.2 {
		section = "????????????: ????????????"
	} else if ahr999x > 1.2 && ahr999x <= 5 {
		section = "????????????: ????????????"
	} else if ahr999x > 5 {
		section = "???????????????????????????"
	}
	datastring := "????????????:" + prices + "\n200???????????????:" + avgs + "\n????????????:" + logprices + "\nAhr999X??????:" + ahr999s + "\n" + section
	return datastring
}

// 30 mins check if hit: send message to me (DONE)
// subscribe module							(db DONE)
// (payment module)(web for sub)

func message() {
	ahr999button := `{"label": "Ahr999", "action": "input:ahr999", "color": "#5979F0"}`
	ahr999xbutton := `{"label": "Ahr999X", "action": "input:ahr999x", "color": "#5979F0"}`
	introbutton := `{"label": "????????????", "action": "input:?", "color": "#5979F0"}`
	subbutton := `{"label": "????????????", "action": "input:/sub", "color": "#5979F0"}`
	unsubbutton := `{"label": "????????????", "action": "input:/unsub", "color": "#B76753"}`
	statusbutton := `{"label": "????????????", "action": "input:/status", "color": "#6BC0CE"}`
	ahr999introbutton := `{"label": "Ahr999????????????", "action": "input:/ahr999intro", "color": "#75A2CB"}`
	//donatebutton := `{"label": "???????????????????????????????", "action": "https://donate.cafe/who3m1", "color":"#0080FF"}`
	helpmessagePost := `
## ???????????????:
Ahr999???????????????????????????????????????????????????ahr999??????????????????????????????????????????????????????????????????

## ??????:
- ahr999	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;???????????????ahr999??????
- ahr999x	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;???????????????ahr999x??????
- /sub	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;??????ahr999????????????(??????????????????)
- /unsub	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;???????????? 
- ?		&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;????????????
- /ahr999intro	&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;??????ahr999???????????????
- ?????????????????????    &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;ahr999????????????

## ?????? 
- [???????????????????????????????](https://donate.cafe/who3m1)
`
	ahr999introPost := `
## Ahr999(??????)??????

<img src="https://ahr999.com/ahr999.jpg" width = "200" height = "50%" alt="ahr999??????" align=center />

Ahr999 ??????????????????????????????????????????????????????????????????????????????????????????????????????

??????????????????2021.08.19??????????????????????????????:[http://btc.mom/?cat=154]()

????????????:[http://cdn.fromfriend.com/HODLBITCOIN_ahr999.pdf]()


## Ahr999 ????????????    

Ahr999?????????????????????@Ahr999?????????????????????????????????????????????????????????????????????

Ahr999?????? = ??????????????????/200?????????????????? * ??????????????????/?????????????????????

- ahr999????????????0.45????????????????????????

- ahr999????????????0.45???1.2??????????????????????????????

- ahr999??????1.2???5????????????????????????????????????

- ahr999????????????5????????????????????????

`
	tunbtclink := "https://cdn.fromfriend.com/HODLBITCOIN_ahr999.pdf"

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

		switch string(data) {
		// for dev
		case "showid":
			covidString := msg.ConversationID
			useridString := msg.UserID
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte(covidString), 2)
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte(useridString), 3)
		case "/display":
			number := displaySubuser(sqliteDatabase)
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte(number), 3)
		// usages
		case "ahr999":
			ahr := getahr999string()
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte(ahr), 4)
		case "ahr999x":
			ahrx := getahr999xstring()
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte(ahrx), 4)
		case "/sub":
			s := checkSubUser(sqliteDatabase, msg.UserID)
			if s {
				controlSub(sqliteDatabase, "true", msg.UserID)
			} else if !s {
				insertSubuser(sqliteDatabase, msg.UserID, msg.ConversationID, true)
			}
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte("????????????????????????????????????????????????????????????"), 5)
			//Mixinrespond(client, ctx, msg, mixin.MessageCategoryAppButtonGroup, []byte(donatebutton), 5)
		case "/unsub":
			controlSub(sqliteDatabase, "false", msg.UserID)
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte("????????????????????????????????????????????????????????????????????????????????????/del??????????????????????????????????????????)"), 6)
		case "/del":
			deleteSubuser(sqliteDatabase, msg.UserID)
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte("??????????????????"), 6)
		case "/status":
			status := statusSubuser(sqliteDatabase, msg.UserID)
			log.Println(status)
			Mixinrespond(client, ctx, msg, "PLAIN_TEXT", []byte(status), 6)

		//help
		case "?", "/?", "???", "/???":
			msgbutton := "[" + ahr999button + "," + ahr999xbutton + "," + subbutton + "," + unsubbutton + "," + statusbutton + "," + ahr999introbutton + "]"
			Mixinrespond(client, ctx, msg, mixin.MessageCategoryPlainPost, []byte(helpmessagePost), 7)
			if err := Mixinrespond(client, ctx, msg, mixin.MessageCategoryAppButtonGroup, []byte(msgbutton), 8); err != nil {
				log.Println(err)
			}
		case "Hi", "??????":
			msgbutton := "[" + introbutton + "," + subbutton + "," + unsubbutton + "," + statusbutton + "," + ahr999introbutton + "]"
			if err := Mixinrespond(client, ctx, msg, mixin.MessageCategoryAppButtonGroup, []byte(msgbutton), 8); err != nil {
				log.Println(err)
			}
		case "/ahr999intro":
			if err := Mixinrespond(client, ctx, msg, mixin.MessageCategoryPlainPost, []byte(ahr999introPost), 8); err != nil {
				log.Println(err)
			}
			Mixinrespond(client, ctx, msg, mixin.MessageCategoryPlainText, []byte("????????????:"+tunbtclink), 7)
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
	var wg sync.WaitGroup
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

	b := func(){
		index := getahr999string()
		goMixinSubBroadcast(sqliteDatabase, client, ctx, []byte(index), &wg)
	}
	c := cron.New()
	c.AddFunc("0 0 * * *", b)
	c.Start()
	// main loop
	for {
		time.Sleep(time.Second * 60)
	}
}
