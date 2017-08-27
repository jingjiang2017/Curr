package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"

	"syscall"
	"time"

	"gopkg.in/robfig/cron.v2"

	"github.com/qiniu/db/mgoutil.v3"
	"github.com/qiniu/log.v1"
	"qbox.us/cc/config"
)

type M map[string]interface{}
type Config struct {
	MaxProcs   int            `json:"max_procs"`
	DebugLevel int            `json:"debug_level"`
	Mgo        mgoutil.Config `json:"mgo"`
	Urls       []string       `json:"urls"`
}

var NowNo = int64(0)
var cfg = Config{
	MaxProcs:   1,
	DebugLevel: 1,
	Mgo:        mgoutil.Config{Host: "127.0.0.1", DB: "racing_dev2"},
	Urls:       []string{"http://d.apiplus.net/newly.do?token=t31ca37cd375be4b4k&code=bjpk10&rows=1&format=json"},
}

func main() {
	config.Init("f", "fetcher", "fetcher.json")
	if err := config.Load(&cfg); err != nil {
		log.Warn("Load fetch config file failed\n Use default config:")
	}
	runtime.GOMAXPROCS(cfg.MaxProcs)
	log.SetOutputLevel(cfg.DebugLevel)
	c := cron.New()
	//c.AddFunc("", func() { fetcher(cfg) })
	//c.AddFunc("0 * * * * ?", testCron)
	c.AddFunc("20 2-59/5 9-23 * * ?", run)
	c.Start()
	//	r, err := fetcher(cfg)
	//	fmt.Printf("%#v\n, %v", r, err)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-sigs
	os.Exit(0)
}

func testCron2() {
	fmt.Printf("==> Run at \t%v \n", time.Now())
}
func run() {
	start := time.Now()
	mgr := NewLotteryMgr(cfg)
	log.Infof(">>>>>>>>start to fetch lottery<<<<<<<<<")
	lt, err := mgr.fetcher()
	if err != nil {
		log.Error(err)
	}
	for time.Unix(lt.Opentimestamp, 0).Minute() != time.Now().Minute() {
		log.Infof("get old lottery record [no:%d], try again", lt.No)
		lt, err = mgr.reTry()
		if err != nil {
			log.Error(err)
		}
	}
	NowNo = lt.No

	if err = mgr.store(lt); err != nil {
		log.Error(err)
		return
	}

	log.Infof("Cost: %v", time.Since(start))
	log.Infof(">>>>>>>>finish fetching lottery<<<<<<<<<")
}

type LotteryMgr struct {
	cfg   Config
	colls Collections
	trys  int
}

func NewLotteryMgr(cfg Config) *LotteryMgr {
	var colls Collections
	_, err := mgoutil.Open(&colls, &cfg.Mgo)
	for err != nil {
		log.Warnf("Open mongodb failed: %v", err)
		_, err = mgoutil.Open(&colls, &cfg.Mgo)
		time.Sleep(time.Second * 1)
	}
	return &LotteryMgr{cfg, colls, 0}
}

type Bet struct {
	No       int64  `json:"no" bson:"no"`
	Nickname string `json:"nickname" bson:"nickname"`
	Openid   string `json:"openid" bson:"openid"`
	Choice   string `json:"choice" bson:"choice"`
	Amount   int    `json:"amount" bson:"amount"`
	Avatar   string `json:"avatar" bson:"avatar"`
}

func (m *LotteryMgr) stat() {
	var bets []Bet
	err := m.colls.BetColl.Find(M{"no": NowNo, "from": 1}).All(&bets)
	if err != nil {
		log.Errorf("failed to get bets, error: %v", err)
	}
	for _, v := range bets {
		fmt.Printf("%#v\n", v)
	}
}

func (m *LotteryMgr) store(lt Lottery) error {
	selector := M{"no": lt.No}
	update := M{"no": lt.No, "code": lt.Opencode, "opentime": lt.Opentime}
	_, err := m.colls.LotteryColl.Upsert(selector, update)
	if err != nil {
		log.Errorf("failed to persist lottery record, error: %v", err)
		return err
	}
	log.Infof("success to get lottery[%#v]", lt)
	return nil
}

func (m *LotteryMgr) reTry() (lt Lottery, err error) {
	m.trys += 1
	log.Infof("retry to fetch lottery %d ...", m.trys)
	time.Sleep(time.Second * 3)
	return m.fetcher()
}

func (m *LotteryMgr) fetcher() (lt Lottery, err error) {
	resp, err2 := http.Get(m.cfg.Urls[0])
	if err2 != nil {
		log.Warnf("try to get lottery failed, error: %v", err2)
		m.reTry()
	}
	if resp != nil {
		defer func() {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	var b []byte
	if b, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Errorf("read resp body error: %v", err)
		m.reTry()
	}
	var lts Lotterys
	if err = json.Unmarshal(b, &lts); err != nil {
		log.Errorf("parse resp body error: %v", err)
		m.reTry()
	}
	lt = lts.Data[0]
	no, err := strconv.ParseInt(lt.Expect, 10, 64)
	if err != nil {
		log.Errorf("failed to convent string to int, except=%s, error: %v", lt.Expect, err)
		return
	}
	lt.No = no
	m.trys = 0
	return
}

type Lottery struct {
	No            int64  `json:"no"`
	Expect        string `json:"expect"`
	Opencode      string `json:"opencode"`
	Opentime      string `json:"opentime"`
	Opentimestamp int64  `json:"opentimestamp"`
}
type Lotterys struct {
	Rows int       `json:"rows"`
	Code string    `json:"code"`
	Data []Lottery `json:"data"`
}
type Collections struct {
	UserColl    mgoutil.Collection `coll:"users"`
	LotteryColl mgoutil.Collection `coll:"lotterys"`
	BetColl     mgoutil.Collection `coll:"bets"`
}