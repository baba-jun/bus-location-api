package handler

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

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

//go:embed assets/odpt_BusstopPole_Toei.json
var toeiData []byte

func Handler(w http.ResponseWriter, r *http.Request) {
	// CORS設定
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

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

	// 現在は都営バスのみサポート
	if operatorName != "Toei" {
		http.Error(w, "Only Toei operator is supported", http.StatusBadRequest)
		return
	}

	// 埋め込まれたデータを使用
	var odptBusstops []ODPTBusstopPole
	if err := json.Unmarshal(toeiData, &odptBusstops); err != nil {
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
