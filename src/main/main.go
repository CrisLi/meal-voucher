package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/go-redis/redis"
	"github.com/mariopinderist/cloudwatch"
	"gitlab.com/modulus.io/gme/config"
	"gitlab.com/modulus.io/gme/logger"
	"gitlab.com/modulus.io/gme/matchingengine"
	"gitlab.com/modulus.io/gme/risk"

	nats "github.com/nats-io/go-nats-streaming"
)

// This will hold the matching engine instances
var matchingEngines map[string]*matchingengine.MatchingEngine

// the creation locks
var matchingEngineLocks map[string]sync.RWMutex

var natsC nats.Conn

var key = "8e6c26ad7e60402d165216ad9c7cd5cbc92be723c0de00768ca61622df77fb1fb472922ae328413be9fb4a1c7bda8ab8619686bb88ac48378a92a0f8599b03bf"
var exchange = "https://demo6admin.modulusexchange.com"
var checkLicence = "true"

var allocatedOrders = 10000

var restartChan chan struct{}

var redisClient *redis.Client = nil
var redisClientOrders *redis.Client = nil

var redisURL string
var redisPass string
var redisAuth bool

var version = "5.4"
var marketOrderMaxSlipPercLB = 0.9
var marketOrderMaxSlipPercUB = 1.1
var specialUsersVM = map[uint64]bool{332525: true}
var specialUsersLP = map[uint64]bool{332424: true}
var vcmEnabled bool = false
var maxOpenOrders = 199

func main() {

	_debug := flag.Bool("d", false, "log the messages to the console")
	_port := flag.Int("p", 8000, "set the listening port")
	_path := flag.String("path", "/tmp/gme", "set the data path for file storage")
	_clusterID := flag.String("cluster", "test-cluster", "the NATS cluster ID")
	_clientID := flag.String("client", "nats_server", "the NATS client ID")
	_nats := flag.Bool("nats", true, "enable NATS, disable HTTP")
	_prof := flag.Bool("prof", false, "enable profile interface")
	_cloudWatch := flag.Bool("aws", false, "write logs to aws")
	_cloudWatchGroup := flag.String("awsgroup", "GME", "CloudWatch group name")
	_cloudWatchKeyID := flag.String("awskeyid", "awskeyid", "AWS Key Id")
	_cloudWatchSecretKey := flag.String("awssecretkey", "awssecretkey", "AWS Secret Key")
	_awsRegion := flag.String("awsregion", "eu-central-1", "AWS Region")
	_exchange := flag.String("exchange", "", "Exchange URL.")
	_key := flag.String("key", "", "Key.")
	_orders := flag.String("orders", "", "Allocated orders in memory")
	_natsURL := flag.String("nats-url", "nats://localhost:4222", "nats url")
	_marketOrderMaxSlipPerc := "10"
	_mongoConnStr := flag.String("mongo-connstr", "mongodb://loalhost:27017", "mongo-connstr")

	_riskFC := flag.Float64("risk-FC", 40, "set the risk tolerence for frequentCancellation")
	_riskPD := flag.Float64("risk-PD", 40, "set the risk tolerence for pumpDump")
	_riskSL := flag.Float64("risk-SL", 0.25, "set the risk tolerence for spoofingLayering")
	_riskEngineMode := flag.String("risk-EngineMode", "", "set the risk EngineMode. `OT` for onTrade `AT` for aftertrade")
	_riskEngineWhitelistedUserIDs := flag.String("risk-EngineWhitelistedUserIDs", "332424,332525", "set the risk Engine Whitelisted UserIDs csv.")

	_redisUrl := flag.String("redis_url", "localhost:6379", "Redis URLs")
	_redisPass := flag.String("redis_pass", "vpdS6HGDa3gx8tUXmJSm5D7yQ9E7fkwj4dt4H38z", "Redis Pass")
	_redisAuth := flag.Bool("redis_auth", true, "Redis Auth")

	_vcmEnabled := flag.Bool("vcm", false, "VCM Enabled")
	_maxOpenOrdersPerUser := flag.Int("max_open_orders_per_user", 200, "Maximum open orders per customer per trading pair, default is 200")

	flag.Parse()

	if os.Getenv("vcm") != "" {
		*_vcmEnabled = os.Getenv("vcm") == "true"
	}

	if os.Getenv("max_open_orders_per_user") != "" {
		_max_open_orders_per_user_env := os.Getenv("max_open_orders_per_user")
		if s, err := strconv.Atoi(_max_open_orders_per_user_env); err == nil {
			*_maxOpenOrdersPerUser = s
		}
	}

	if os.Getenv("max-slip-perc") != "" {
		_marketOrderMaxSlipPerc = os.Getenv("max-slip-perc")
	}

	if _marketOrderMaxSlipPerc != "10" {
		if s, err := strconv.ParseFloat(_marketOrderMaxSlipPerc, 32); err == nil {
			if s > 0 {
				marketOrderMaxSlipPercUB = 1 + (s * 0.01)
				marketOrderMaxSlipPercLB = 1 - (s * 0.01)
			}
		}
	}

	if *_exchange != "" {
		exchange = *_exchange
	}
	if *_key != "" {
		key = *_key
	}

	if os.Getenv("redis_pass") != "" {
		*_redisPass = os.Getenv("redis_pass")
	}

	if os.Getenv("redis_url") != "" {
		*_redisUrl = os.Getenv("redis_url")
	}

	if os.Getenv("redis_auth") != "" {
		*_redisAuth = os.Getenv("redis_auth") == "true"
	}

	if os.Getenv("mongo-connstr") != "" {
		*_mongoConnStr = os.Getenv("mongo-connstr")
	}

	if os.Getenv("risk-FC") != "" {
		*_riskFC, _ = strconv.ParseFloat(os.Getenv("risk-FC"), 64)
	}

	if os.Getenv("risk-PD") != "" {
		*_riskPD, _ = strconv.ParseFloat(os.Getenv("risk-PD"), 64)
	}

	if os.Getenv("risk-SL") != "" {
		*_riskSL, _ = strconv.ParseFloat(os.Getenv("risk-SL"), 64)
	}

	if os.Getenv("risk-EngineMode") != "" {
		*_riskEngineMode = strings.ToUpper(strings.TrimSpace(os.Getenv("risk-EngineMode")))
	}

	if os.Getenv("risk-EngineWhitelistedUserIDs") != "" {
		*_riskEngineWhitelistedUserIDs = os.Getenv("risk-EngineWhitelistedUserIDs")
	}

	maxOpenOrders = *_maxOpenOrdersPerUser
	natsURL := *_natsURL
	redisURL = *_redisUrl
	redisPass = *_redisPass
	redisAuth = *_redisAuth

	if !redisAuth {
		redisClient = redis.NewClient(&redis.Options{
			Addr: *_redisUrl,
			DB:   1,
		})
	} else {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     *_redisUrl,
			Password: *_redisPass,
			DB:       1,
		})
	}

	_, err := redisClient.Ping().Result()
	if err != nil {
		log.Print(err)
	}

	if tempNats := os.Getenv("nats-url"); tempNats != "" {
		natsURL = tempNats
	}
	if tempNats := os.Getenv("nats-cluster"); tempNats != "" {
		*_clusterID = tempNats
	}
	if tempNats := os.Getenv("nats-client"); tempNats != "" {
		*_clientID = tempNats
	}
	if tempNats := os.Getenv("exchange"); tempNats != "" {
		exchange = tempNats
	}
	if tempNats := os.Getenv("key"); tempNats != "" {
		key = tempNats
	}
	if *_orders != "" {
		o, err := strconv.Atoi(*_orders)
		if err != nil {
			log.Println("can not set allocated orders", err)
		} else {
			allocatedOrders = o
		}
	}
	fmt.Println(exchange)

	vcmEnabled = *_vcmEnabled

	// Insert logs to AWS CloudWatch.
	if *_cloudWatch {
		os.Setenv("AWS_ACCESS_KEY_ID", *_cloudWatchKeyID)
		os.Setenv("AWS_SECRET_ACCESS_KEY", *_cloudWatchSecretKey)
		awsSession, err := session.NewSession(&aws.Config{Region: aws.String(*_awsRegion)})
		if err != nil {
			log.Fatal(err)
		}
		client := cloudwatchlogs.New(awsSession)
		group := cloudwatch.NewGroup(*_cloudWatchGroup, client)
		w, err := group.Create(*_cloudWatchGroup + "-" + fmt.Sprint(time.Now().UnixNano()))
		if err != nil {
			log.Println("can not log to aws: ", err)
		} else {
			writer := io.MultiWriter(w, os.Stdout)
			log.SetOutput(writer)
		}
	}

	restartChan = make(chan struct{})

	go func() {
		select {
		case <-restartChan:
			panic("restart requested")
		}
	}()

	config.SetDataDirectory(*_path)

	logger.DebugLevel = *_debug

	port := *_port

	// Pass traffic to NATS.
	if *_nats {

		logger.Info("NATS enabled, cluster %s, client %s, redis %s (auth:%s), version %s\n", *_clusterID, *_clientID, redisURL, redisAuth, version)
		logger.Info("Market Order Protection Enabled  LB = %f, UB = %f\n", marketOrderMaxSlipPercLB, marketOrderMaxSlipPercUB)
		logger.Info("Max Open Orders per customer per trading pair is %d\n", maxOpenOrders)

		if vcmEnabled {
			logger.Info("VCM Enabled with (Min,Max) = (0,0), VCM will work only when at least one of the (Min,Max) is non zero \n")
		}
		var err error
		natsC, err = nats.Connect(*_clusterID, *_clientID, nats.NatsURL(natsURL))
		if err != nil {
			panic(err)
		}

		go func() {
			ticker := time.NewTicker(500 * time.Millisecond)
			for {
				select {
				case <-ticker.C:
					if natsC == nil || !natsC.NatsConn().IsConnected() {
						if natsC != nil {
							natsC.Close()
						}
						natsC, err = nats.Connect(*_clusterID, *_clientID, nats.NatsURL(natsURL))
						if err != nil {
							log.Println(err)
						} else {
							log.Println("nats reconected")
							for _, m := range matchingEngines {
								m.Kill(false)
								err = m.ReSub(natsC)
								if err != nil {
									log.Println(err)
								}
							}
							log.Println("resub finished")
						}
					}
				}
			}
		}()
	} else {
		logger.Info("HTTP enabled on port %d\n", port)
	}

	go func () {
		for {
			now := time.Now().UTC()
			tomorrow := now.AddDate(0, 0, 1)
			next := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC)
			diff := next.Sub(now)
			log.Println("cancel daily order : starts in ", diff)
			select {
			case <-time.Tick(diff):
				for _, m := range matchingEngines {
					m.CancelDailyOrders()
				}
			}
		}
	}()

	if !redisAuth {
		redisClientOrders = redis.NewClient(&redis.Options{
			Addr: redisURL,
			DB:   11,
		})
		_, err = redisClientOrders.Ping().Result()
		if err != nil {
			log.Print("redisClientOrders.Ping().Result(): ", err)
			panic(err)
		}

		err = redisClientOrders.FlushDB().Err()
		if err != nil {
			log.Print("redisClientOrders.FlushDB().Err(): ", err)
			panic(err)
		}
	} else {
		redisClientOrders = redis.NewClient(&redis.Options{
			Addr:     redisURL,
			Password: redisPass,
			DB:       11,
		})
		_, err = redisClientOrders.Ping().Result()
		if err != nil {
			log.Print("redisClientOrders.Ping().Result(): ", err)
			panic(err)
		}

		err = redisClientOrders.FlushDB().Err()
		if err != nil {
			log.Print("redisClientOrders.FlushDB().Err(): ", err)
			panic(err)
		}
	}

	// Prep the matching engines map
	matchingEngines = make(map[string]*matchingengine.MatchingEngine, 0)

	if checkLicence == "true" {
		//Check Licensing with exchange.
		var i int = 1
		for {
			logger.Info("\n%d checking license %s", i, exchange)
			err := checkLicensingWithExchange(exchange)
			if err != nil {
				logger.Info(" ... error checking license: %v", err)
				if i > 10 {
					if natsC != nil {
						natsC.Close()
					}

					panic(err)
				} else {
					i++
					time.Sleep(10 * time.Second)
				}

			} else {
				logger.Info("\nlicense verified\n")
				break
			}
		}
	}

	//Update Trade history
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for {
			select {
			case <-ticker.C:
				updateTrades()
			}
		}
	}()

	if *_riskEngineMode == "ONTRADE" || *_riskEngineMode == "POSTTRADE" {
		err = risk.ConnectToMongoDB(*_mongoConnStr)
		if err == nil {
			risk.SetRiskEngineParameters(_riskFC, _riskPD, _riskSL, _riskEngineMode, _riskEngineWhitelistedUserIDs)
		} else {
			fmt.Println("MongoDB connection failed", err)
		}
		fmt.Println("RISK-Engine Mode ", *_riskEngineMode)
	}
	//else {
	//	fmt.Println("Invalid risk-EngineMode ", *_riskEngineMode, ". Supported values are `ONTRADE` for on-trade and `POSTTRADE` for after-trade")
	//}

	r := http.NewServeMux()

	r.HandleFunc("/RiskEngineStatus", func(writer http.ResponseWriter, request *http.Request) {
		outputSuccessfulAPIResponse(writer, []byte(`{"RiskEngineMode": "`+*_riskEngineMode+`","RiskFC" : "`+fmt.Sprint(*_riskFC)+`","RiskPD" : "`+fmt.Sprint(*_riskPD)+`","RiskSL" : "`+fmt.Sprint(*_riskSL)+`","WhitelistedUserIDs" : "`+fmt.Sprint(*_riskEngineWhitelistedUserIDs)+`"}`))
	})
	// Associate the http handlers
	r.HandleFunc("/StartMatchingEngine", startMatchingEngineHandler)
	r.HandleFunc("/StopMatchingEngine", killMatchingEngineHandler)
	r.HandleFunc("/VCM", vcmHandler)
	r.HandleFunc("/healthcheck", func(writer http.ResponseWriter, request *http.Request) {})
	r.HandleFunc("/forcerestart", func(writer http.ResponseWriter, request *http.Request) {
		restartChan <- struct{}{}
	})
	r.HandleFunc("/version", func(writer http.ResponseWriter, request *http.Request) {
		outputSuccessfulAPIResponse(writer, []byte(`{"version": "`+version+`"}`))
	})
	r.HandleFunc("/restart", func(writer http.ResponseWriter, request *http.Request) {
		log.Println("restart request received.")
		// Get Message and CipherMessage from headers.
		if checkLicence == "true" {
			message := request.Header.Get("Message") + "7da79d73-71dc-47d0-b317-4d299c90ad10"
			cipherMessage := request.Header.Get("CipherMessage")
			// Check if Message or CipherMessage are empty.
			if message == "" || cipherMessage == "" {
				outputErrorAPIJsonResponse(writer, errors.New("licensing problems"), http.StatusBadRequest)
				return
			}
			//Generate hash from Message.
			sha512 := hmac.New(sha512.New, []byte(key))
			sha512.Write([]byte(message))
			// Compare CipherMessage with hashed message.
			if strings.Replace(strings.ToUpper(hex.EncodeToString(sha512.Sum(nil))), "-", "", -1) != cipherMessage {
				outputErrorAPIJsonResponse(writer, errors.New("licensing problems"), http.StatusBadRequest)
				return
			}
		}
		//restartChan <- struct{}{}
		for n, m := range matchingEngines {
			m.Kill(true)
			delete(matchingEngines, n)
		}
		outputSuccessfulAPIResponse(writer, []byte("success"))
	})
	r.HandleFunc("/restartnats", func(writer http.ResponseWriter, request *http.Request) {
		// Get Message and CipherMessage from headers.
		message := request.Header.Get("Message") + "7da79d73-71dc-47d0-b317-4d299c90ad10"
		cipherMessage := request.Header.Get("CipherMessage")
		// Check if Message or CipherMessage are empty.
		if message == "" || cipherMessage == "" {
			outputErrorAPIJsonResponse(writer, errors.New("licensing problems"), http.StatusBadRequest)
			return
		}
		//Generate hash from Message.
		sha512 := hmac.New(sha512.New, []byte(key))
		sha512.Write([]byte(message))
		// Compare CipherMessage with hashed message.
		if strings.Replace(strings.ToUpper(hex.EncodeToString(sha512.Sum(nil))), "-", "", -1) != cipherMessage {
			outputErrorAPIJsonResponse(writer, errors.New("licensing problems"), http.StatusBadRequest)
			return
		}
		for n, m := range matchingEngines {
			m.Kill(true)
			delete(matchingEngines, n)
		}

		cmd := exec.Command("/bin/sh", "/home/ubuntu/restart_nats.sh")
		err := cmd.Run()
		if err != nil {
			log.Printf("cmd.Run() failed with %s\n", err)
		}
		//natsC, err = nats.Connect(*_clusterID, *_clientID)
		//if err != nil {
		//	log.Println("NATS error: ", err)
		//}
		////time.Sleep(3*time.Second)
		restartChan <- struct{}{}
		outputSuccessfulAPIResponse(writer, []byte("success"))
	})

	r.HandleFunc("/SubmitOrder", submitOrderHandler)
	r.HandleFunc("/CancelOrder", cancelOrderHandler)
	r.HandleFunc("/CancelPricePoints", cancelPricePointHandler)

	if *_prof {
		// Register pprof handlers
		r.HandleFunc("/debug/pprof/", pprof.Index)
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		r.HandleFunc("/debug/pprof/profile", pprof.Profile)
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		r.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	r.HandleFunc("/Book", bookHandler)
	r.HandleFunc("/LastTrades", lastTradesHandler)
	r.HandleFunc("/Status", statusHandler)
	r.HandleFunc("/monitor", monitorHandler)
	r.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))

	// TEMP
	r.HandleFunc("/PrettyPrint", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("PrettyPrint API call received.\n")
		w.WriteHeader(http.StatusOK)
		for _, m := range matchingEngines {
			w.Write([]byte(m.PrettyPrint()))
		}
	})

	//this is for reproducing crash counter closed bug
	r.HandleFunc("/kill", func(writer http.ResponseWriter, request *http.Request) {

		outputSuccessfulAPIResponse(writer, []byte(`{"message": "removed"}`))
		return

		currencyPair := request.URL.Query().Get("pair")

		if len(currencyPair) < 2 {
			outputSuccessfulAPIResponse(writer, []byte(`{"message": "no quer parameter pair supplied"}`))
			return
		}

		// Find the right matching engine
		var (
			me *matchingengine.MatchingEngine
			ok bool
		)
		if me, ok = matchingEngines[currencyPair]; !ok {
			outputSuccessfulAPIResponse(writer, []byte(`{"message": "invalid quer parameter pair"}`))
			return
		}
		me.Kill(true)
		outputSuccessfulAPIResponse(writer, []byte(`{"message": "successfully killed pair :"`+currencyPair+` "}`))
	})

	r.HandleFunc("/specialusers", func(w http.ResponseWriter, r *http.Request) {

		// Log the API call.
		//if logger.DebugLevel {
		logger.Debug("specialusers : API call received.\n")
		//}

		reqBodyBytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Info("specialusers : ReadAll err :", err)
			outputErrorAPIJsonResponse(w, err, http.StatusBadRequest)
			return
		}

		splUser := &SpecialUser{}
		err = json.Unmarshal(reqBodyBytes, splUser)
		if err != nil {
			logger.Info("specialusers : Unmarshal err :", err)
			outputErrorAPIJsonResponse(w, err, http.StatusBadRequest)
			return
		}

		if splUser == nil || len(splUser.UserIDs) == 0 {
			logger.Info("specialusers : invalid payload :", string(reqBodyBytes))
			outputErrorAPIJsonResponse(w, err, http.StatusBadRequest)
			return
		}

		// for k := range specialUsers {
		// 	if k == 332424 || k == 332525 {
		// 		continue
		// 	}
		// 	delete(specialUsers, k)
		// }

		// for _, usrid := range splUser.UserIDs {
		// 	specialUsers[usrid] = true
		// }

		// allSplUser := &SpecialUser{}
		// for usrid := range specialUsers {
		// 	allSplUser.UserIDs = append(allSplUser.UserIDs, usrid)
		// }

		allSplUser := PopulateSpecialUsersLP(splUser.UserIDs)
		response, err := json.Marshal(&allSplUser)
		if err != nil {
			outputErrorAPIJsonResponse(w, err, http.StatusInternalServerError)
			return
		}
		outputSuccessfulAPIResponse(w, response)

	})

	go fetchAndPopulateSpecialUsersLP(exchange, &specialUsersLP)

	// Start the http server
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), r))

}
