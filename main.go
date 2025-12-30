package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Bus レスポンスの構造体
type Bus struct {
	ID                  string     `json:"id"`
	Type                string     `json:"type"`
	Date                time.Time  `json:"date"`
	Note                string     `json:"note"`
	Operator            string     `json:"operator"`
	BusNumber           string     `json:"busNumber"`
	BusTimetable        string     `json:"busTimetable,omitempty"`
	ToBusstopPole       string     `json:"toBusstopPole,omitempty"`
	BusroutePattern     string     `json:"busroutePattern,omitempty"`
	FromBusstopPole     string     `json:"fromBusstopPole,omitempty"`
	FromBusstopPoleTime *time.Time `json:"fromBusstopPoleTime,omitempty"`
	StartingBusstopPole string     `json:"startingBusstopPole,omitempty"`
	TerminalBusstopPole string     `json:"terminalBusstopPole,omitempty"`
}

// ODPTのレスポンス構造体
type ODPTBus struct {
	ID                  string `json:"@id"`
	Type                string `json:"@type"`
	Date                string `json:"dc:date"`
	Context             string `json:"@context"`
	Valid               string `json:"dct:valid"`
	Note                string `json:"odpt:note"`
	SameAs              string `json:"owl:sameAs"`
	Busroute            string `json:"odpt:busroute"`
	Operator            string `json:"odpt:operator"`
	BusNumber           string `json:"odpt:busNumber"`
	Frequency           int    `json:"odpt:frequency"`
	BusTimetable        string `json:"odpt:busTimetable"`
	ToBusstopPole       string `json:"odpt:toBusstopPole"`
	BusroutePattern     string `json:"odpt:busroutePattern"`
	FromBusstopPole     string `json:"odpt:fromBusstopPole"`
	FromBusstopPoleTime string `json:"odpt:fromBusstopPoleTime"`
	StartingBusstopPole string `json:"odpt:startingBusstopPole"`
	TerminalBusstopPole string `json:"odpt:terminalBusstopPole"`
}

// BusstopPole レスポンスの構造体
type BusstopPole struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	SameAs   string   `json:"sameAs"`
	Date     string   `json:"date"`
	Title    string   `json:"title"`
	Long     float64  `json:"long"`
	Lat      float64  `json:"lat"`
	Operator []string `json:"operator"`
}

// ODPTのバス停データ構造体
type ODPTBusstopPole struct {
	ID       string      `json:"@id"`
	Type     string      `json:"@type"`
	Title    interface{} `json:"title"`
	Date     string      `json:"dc:date"`
	DCTitle  string      `json:"dc:title"`
	Long     float64     `json:"geo:long"`
	Lat      float64     `json:"geo:lat"`
	Kana     string      `json:"odpt:kana"`
	Note     string      `json:"odpt:note"`
	SameAs   string      `json:"owl:sameAs"`
	Operator []string    `json:"odpt:operator"`
}

const odptAPIBaseURL = "https://api-public.odpt.org/api/v4"

// CORSミドルウェア
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// バス位置情報を取得するハンドラー
func getBusVehicleLocation(w http.ResponseWriter, r *http.Request) {
	// クエリパラメータからoperatorを取得
	operator := r.URL.Query().Get("operator")

	if operator == "" {
		http.Error(w, "operator parameter is required", http.StatusBadRequest)
		return
	}

	// オプションのフィルタパラメータを取得
	busNumber := r.URL.Query().Get("busNumber")
	busTimetable := r.URL.Query().Get("busTimetable")
	toBusstopPole := r.URL.Query().Get("toBusstopPole")
	busroutePattern := r.URL.Query().Get("busroutePattern")
	fromBusstopPole := r.URL.Query().Get("fromBusstopPole")
	startingBusstopPole := r.URL.Query().Get("startingBusstopPole")
	terminalBusstopPole := r.URL.Query().Get("terminalBusstopPole")

	// ODPT APIにリクエストを送信
	apiURL := fmt.Sprintf("%s/odpt:Bus", odptAPIBaseURL)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// パラメータを設定
	q := url.Values{}
	q.Add("odpt:operator", operator)

	// オプションパラメータを追加
	if busNumber != "" {
		q.Add("odpt:busNumber", busNumber)
	}
	if busTimetable != "" {
		q.Add("odpt:busTimetable", busTimetable)
	}
	if toBusstopPole != "" {
		q.Add("odpt:toBusstopPole", toBusstopPole)
	}
	if busroutePattern != "" {
		q.Add("odpt:busroutePattern", busroutePattern)
	}
	if fromBusstopPole != "" {
		q.Add("odpt:fromBusstopPole", fromBusstopPole)
	}
	if startingBusstopPole != "" {
		q.Add("odpt:startingBusstopPole", startingBusstopPole)
	}
	if terminalBusstopPole != "" {
		q.Add("odpt:terminalBusstopPole", terminalBusstopPole)
	}

	// 環境変数からコンシューマーキーを取得
	consumerKey := os.Getenv("ODPT_CONSUMER_KEY")
	if consumerKey != "" {
		q.Add("acl:consumerKey", consumerKey)
	}

	req.URL.RawQuery = q.Encode()

	log.Printf("Requesting: %s", req.URL.String())

	// リクエストを実行
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error requesting ODPT API: %v", err)
		http.Error(w, "Error requesting external API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ODPT API returned status: %d", resp.StatusCode)
		http.Error(w, fmt.Sprintf("External API returned status: %d", resp.StatusCode), resp.StatusCode)
		return
	}

	// レスポンスを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	// ODPTのレスポンスをパース
	var odptBuses []ODPTBus
	if err := json.Unmarshal(body, &odptBuses); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		http.Error(w, "Error parsing response", http.StatusInternalServerError)
		return
	}

	// ラッパーAPIのレスポンス形式に変換
	buses := make([]Bus, 0, len(odptBuses))
	for _, odptBus := range odptBuses {
		bus := Bus{
			ID:                  odptBus.ID,
			Type:                odptBus.Type,
			Note:                odptBus.Note,
			Operator:            odptBus.Operator,
			BusNumber:           odptBus.BusNumber,
			BusTimetable:        odptBus.BusTimetable,
			ToBusstopPole:       odptBus.ToBusstopPole,
			BusroutePattern:     odptBus.BusroutePattern,
			FromBusstopPole:     odptBus.FromBusstopPole,
			StartingBusstopPole: odptBus.StartingBusstopPole,
			TerminalBusstopPole: odptBus.TerminalBusstopPole,
		}

		// 日時をパース
		if parsedDate, err := time.Parse(time.RFC3339, odptBus.Date); err == nil {
			bus.Date = parsedDate
		}

		if odptBus.FromBusstopPoleTime != "" {
			if parsedTime, err := time.Parse(time.RFC3339, odptBus.FromBusstopPoleTime); err == nil {
				bus.FromBusstopPoleTime = &parsedTime
			}
		}

		buses = append(buses, bus)
	}

	// JSONレスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(buses); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully returned %d bus records for operator: %s", len(buses), operator)
}

// バス停情報を取得するハンドラー
func getBusstopPole(w http.ResponseWriter, r *http.Request) {
	// クエリパラメータからoperatorを取得
	operator := r.URL.Query().Get("operator")

	if operator == "" {
		http.Error(w, "operator parameter is required", http.StatusBadRequest)
		return
	}

	// オプションのフィルタパラメータを取得
	filterID := r.URL.Query().Get("id")
	filterTitle := r.URL.Query().Get("title")
	filterSameAs := r.URL.Query().Get("sameAs")

	// operatorから事業者名を抽出 (例: odpt.Operator:Toei -> Toei)
	operatorParts := strings.Split(operator, ":")
	if len(operatorParts) != 2 {
		http.Error(w, "invalid operator format", http.StatusBadRequest)
		return
	}
	operatorName := operatorParts[1]

	// JSONファイルのパスを構築
	execPath, err := os.Executable()
	if err != nil {
		log.Printf("Error getting executable path: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	execDir := filepath.Dir(execPath)
	filePath := filepath.Join(execDir, "assets", fmt.Sprintf("odpt_BusstopPole_%s.json", operatorName))

	// カレントディレクトリからの相対パスも試す（開発時用）
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		filePath = filepath.Join("assets", fmt.Sprintf("odpt_BusstopPole_%s.json", operatorName))
	}

	log.Printf("Loading busstop data from: %s", filePath)

	// JSONファイルを読み込む
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		http.Error(w, fmt.Sprintf("Data not found for operator: %s", operator), http.StatusNotFound)
		return
	}

	// JSONをパース
	var odptBusstops []ODPTBusstopPole
	if err := json.Unmarshal(file, &odptBusstops); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		http.Error(w, "Error parsing data", http.StatusInternalServerError)
		return
	}

	// ラッパーAPIのレスポンス形式に変換とフィルタリング
	busstops := make([]BusstopPole, 0, len(odptBusstops))
	for _, odptBusstop := range odptBusstops {
		// titleを文字列に変換
		titleStr := ""
		if odptBusstop.DCTitle != "" {
			titleStr = odptBusstop.DCTitle
		} else if titleMap, ok := odptBusstop.Title.(map[string]interface{}); ok {
			if ja, ok := titleMap["ja"].(string); ok {
				titleStr = ja
			}
		} else if titleString, ok := odptBusstop.Title.(string); ok {
			titleStr = titleString
		}

		// フィルタリング処理
		if filterID != "" && odptBusstop.ID != filterID {
			continue
		}
		if filterTitle != "" && !strings.Contains(titleStr, filterTitle) {
			continue
		}
		if filterSameAs != "" && odptBusstop.SameAs != filterSameAs {
			continue
		}

		busstop := BusstopPole{
			ID:       odptBusstop.ID,
			Type:     odptBusstop.Type,
			SameAs:   odptBusstop.SameAs,
			Date:     odptBusstop.Date,
			Title:    titleStr,
			Long:     odptBusstop.Long,
			Lat:      odptBusstop.Lat,
			Operator: odptBusstop.Operator,
		}

		busstops = append(busstops, busstop)
	}

	// JSONレスポンスを返す
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(busstops); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully returned %d busstop records for operator: %s", len(busstops), operator)
}

func main() {
	http.HandleFunc("/location/busvehicle", corsMiddleware(getBusVehicleLocation))
	http.HandleFunc("/busstoppole", corsMiddleware(getBusstopPole))

	log.Println("Starting server on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
