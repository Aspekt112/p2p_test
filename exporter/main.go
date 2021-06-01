package main

import (
  "net/http"
  "log"
  "flag"
  "encoding/json"
  "fmt"
  "crypto/tls"
  "time"
  "strconv"

  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "comsos_testnet"
const statusPage = "/status"
const netInfoPage = "/net_info"

var (
  listenAddress = flag.String("listen-address", ":8080", "Address to listen on for telemetry")
  metricsPath = flag.String("metrics-path", "/metrics", "Path under which to expose metrics")

  tr = &http.Transport{
    TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
  }
  client = &http.Client{Transport: tr}


  currentBlockNumber = prometheus.NewDesc(
    prometheus.BuildFQName(namespace, "", "current_block_number"),
    "Current block number", []string{"channel"}, nil,
  )
  deSync = prometheus.NewDesc(
    prometheus.BuildFQName(namespace, "", "desynchronization_in_seconds"),
    "Desynchronization in seconds (the current time in seconds minus the block time in seconds)", []string{"channel"}, nil,
  )
  peersNumber = prometheus.NewDesc(
    prometheus.BuildFQName(namespace, "", "peers_number"),
    "Current number of peers", []string{"channel"}, nil,
  )
)

type Status struct {
  Jsonrpc string `json:"jsonrpc"`
  ID      int    `json:"id"`
  Result  struct {
    NodeInfo struct {
      ProtocolVersion struct {
        P2P   string `json:"p2p"`
        Block string `json:"block"`
        App   string `json:"app"`
      } `json:"protocol_version"`
      ID         string `json:"id"`
      ListenAddr string `json:"listen_addr"`
      Network    string `json:"network"`
      Version    string `json:"version"`
      Channels   string `json:"channels"`
      Moniker    string `json:"moniker"`
      Other      struct {
        TxIndex    string `json:"tx_index"`
        RPCAddress string `json:"rpc_address"`
      } `json:"other"`
    } `json:"node_info"`
    SyncInfo struct {
      LatestBlockHash     string    `json:"latest_block_hash"`
      LatestAppHash       string    `json:"latest_app_hash"`
      LatestBlockHeight   string    `json:"latest_block_height"`
      LatestBlockTime     time.Time `json:"latest_block_time"`
      EarliestBlockHash   string    `json:"earliest_block_hash"`
      EarliestAppHash     string    `json:"earliest_app_hash"`
      EarliestBlockHeight string    `json:"earliest_block_height"`
      EarliestBlockTime   time.Time `json:"earliest_block_time"`
      CatchingUp          bool      `json:"catching_up"`
    } `json:"sync_info"`
    ValidatorInfo struct {
      Address string `json:"address"`
      PubKey  struct {
        Type  string `json:"type"`
        Value string `json:"value"`
      } `json:"pub_key"`
      VotingPower string `json:"voting_power"`
    } `json:"validator_info"`
  } `json:"result"`
}

type NetInfo struct {
  Jsonrpc string `json:"jsonrpc"`
  ID      int    `json:"id"`
  Result  struct {
    Listening bool     `json:"listening"`
    Listeners []string `json:"listeners"`
    NPeers    string   `json:"n_peers"`
    Peers     []struct {
      NodeInfo struct {
        ProtocolVersion struct {
          P2P   string `json:"p2p"`
          Block string `json:"block"`
          App   string `json:"app"`
        } `json:"protocol_version"`
        ID         string `json:"id"`
        ListenAddr string `json:"listen_addr"`
        Network    string `json:"network"`
        Version    string `json:"version"`
        Channels   string `json:"channels"`
        Moniker    string `json:"moniker"`
        Other      struct {
          TxIndex    string `json:"tx_index"`
          RPCAddress string `json:"rpc_address"`
        } `json:"other"`
      } `json:"node_info"`
      IsOutbound       bool `json:"is_outbound"`
      ConnectionStatus struct {
        Duration    string `json:"Duration"`
        Sendmonitor struct {
          Start    time.Time `json:"Start"`
          Bytes    string    `json:"Bytes"`
          Samples  string    `json:"Samples"`
          Instrate string    `json:"InstRate"`
          Currate  string    `json:"CurRate"`
          Avgrate  string    `json:"AvgRate"`
          Peakrate string    `json:"PeakRate"`
          Bytesrem string    `json:"BytesRem"`
          Duration string    `json:"Duration"`
          Idle     string    `json:"Idle"`
          Timerem  string    `json:"TimeRem"`
          Progress int       `json:"Progress"`
          Active   bool      `json:"Active"`
        } `json:"SendMonitor"`
        Recvmonitor struct {
          Start    time.Time `json:"Start"`
          Bytes    string    `json:"Bytes"`
          Samples  string    `json:"Samples"`
          Instrate string    `json:"InstRate"`
          Currate  string    `json:"CurRate"`
          Avgrate  string    `json:"AvgRate"`
          Peakrate string    `json:"PeakRate"`
          Bytesrem string    `json:"BytesRem"`
          Duration string    `json:"Duration"`
          Idle     string    `json:"Idle"`
          Timerem  string    `json:"TimeRem"`
          Progress int       `json:"Progress"`
          Active   bool      `json:"Active"`
        } `json:"RecvMonitor"`
        Channels []struct {
          ID                int    `json:"ID"`
          Sendqueuecapacity string `json:"SendQueueCapacity"`
          Sendqueuesize     string `json:"SendQueueSize"`
          Priority          string `json:"Priority"`
          Recentlysent      string `json:"RecentlySent"`
        } `json:"Channels"`
      } `json:"connection_status"`
      RemoteIP string `json:"remote_ip"`
    } `json:"peers"`
  } `json:"result"`
}

type Exporter struct {
  Endpoint string
}

func NewExporter(endpoint string) *Exporter {
  return &Exporter{
    Endpoint: endpoint,
  }
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
  ch <- currentBlockNumber  // текущий номер блока
  ch <- deSync              // рассинхрон времени этого блока в секундах (текущее время в секундах минус время блока в секундах, чтобы было видно здоровье этой ноды)
  ch <- peersNumber         // количество пиров
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
  ///////////////////////////////////////////////
  //                  Status
  ///////////////////////////////////////////////
  req, err := http.NewRequest("GET", e.Endpoint + statusPage, nil)
  if err != nil {
    log.Fatal(err)
  }

  resp, err := client.Do(req)
  if err != nil {
    log.Fatal(err)
  }

  status := new(Status)


  err = json.NewDecoder(resp.Body).Decode(&status)
  if err != nil {
    log.Fatal(err)
  }
  latestBlockHeight, _ := strconv.ParseFloat(status.Result.SyncInfo.LatestBlockHeight, 64)
  ch <- prometheus.MustNewConstMetric(
    currentBlockNumber, prometheus.GaugeValue, latestBlockHeight, "cosmos-testnet",
  )
  log.Println("Latest block: " + status.Result.SyncInfo.LatestBlockHeight)


  now := time.Now()
  deSyncInSec := now.Sub(status.Result.SyncInfo.LatestBlockTime).Seconds()
  if err != nil {
    log.Fatal(err)
  }
  ch <- prometheus.MustNewConstMetric(
    deSync, prometheus.GaugeValue, deSyncInSec, "cosmos-testnet",
  )
  log.Println(fmt.Sprintf("Desynchronization: %f sec", deSyncInSec))


  ///////////////////////////////////////////////
  //                  Net Info
  ///////////////////////////////////////////////
  reqt, err := http.NewRequest("GET", e.Endpoint + netInfoPage, nil)
  if err != nil {
    log.Fatal(err)
  }

  // Make request
  resps, err := client.Do(reqt)
  if err != nil {
    log.Fatal(err)
  }

  netInfo := new(NetInfo)

  err = json.NewDecoder(resps.Body).Decode(&netInfo)
  if err != nil {
    log.Fatal(err)
  }
  nPeers, _ := strconv.ParseFloat(netInfo.Result.NPeers, 64)
  ch <- prometheus.MustNewConstMetric(
    peersNumber, prometheus.GaugeValue, nPeers, "cosmos-testnet",
  )
  log.Println(fmt.Sprintf("Peers number: %f", nPeers))


  log.Println("---------------")
}


func main() {
  flag.Parse()

  exporter := NewExporter("http://localhost:26657")
  log.Println("Export metrics from http://localhost:26657")
  prometheus.MustRegister(exporter)

  http.Handle(*metricsPath, promhttp.Handler())
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte(`<html>
                  <head><title>P2P tes: Comsos blockchain exporter</title></head>
                  <body>
                  <h1>Comsos blockchain exporter</h1>
                  <p><a href='` + *metricsPath + `'>Metrics</a></p>
                  </body>
                  </html>`))
  })

  log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
