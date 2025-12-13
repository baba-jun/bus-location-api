# Transport Realtime API Backend

ODPT (Open Data Challenge for Public Transportation) APIをラップするバックエンドサーバー

## 概要

このAPIサーバーは、ODPT APIの公共交通機関データにアクセスするためのラッパーAPIを提供します。

## セットアップ

### 前提条件

- Go 1.21以上
- ODPT APIのコンシューマーキー（[ODPT Developer Site](https://developer.odpt.org/)から取得）

### インストール

```bash
go mod download
```

### 環境変数の設定

ODPT APIを使用するには、コンシューマーキーが必要です。環境変数 `ODPT_CONSUMER_KEY` を設定してください。

#### Windows (PowerShell)

```powershell
$env:ODPT_CONSUMER_KEY="your_consumer_key_here"
```

#### Windows (コマンドプロンプト)

```cmd
set ODPT_CONSUMER_KEY=your_consumer_key_here
```

#### Linux/Mac

```bash
export ODPT_CONSUMER_KEY=your_consumer_key_here
```

または、`.env`ファイルを作成することもできます（将来的な拡張用）。

## 起動方法

```bash
go run main.go
```

または環境変数を設定しながら起動：

```powershell
$env:ODPT_CONSUMER_KEY="your_key"; go run main.go
```

サーバーは `http://localhost:8081` で起動します。

## API エンドポイント

### GET /location/busvehicle

バスの位置情報を取得します。

#### パラメータ

- `operator` (必須): 事業者のID（例: `odpt.Operator:Toei`）

#### リクエスト例

```bash
curl "http://localhost:8081/location/busvehicle?operator=odpt.Operator:Toei"
```

#### レスポンス例

```json
[
  {
    "id": "urn:ucode:_00001C000000000000010000031008D6",
    "type": "odpt:Bus",
    "date": "2025-12-01T17:56:31+09:00",
    "operator": "odpt.Operator:Toei",
    "busNumber": "B786",
    "busTimetable": "odpt.BusTimetable:Toei.RH01.08403-1-09-170-1749",
    "toBusstopPole": "odpt.BusstopPole:Toei.AoyamagakuinChutobu.7.1",
    "busroutePattern": "odpt.BusroutePattern:Toei.RH01.8403.1",
    "fromBusstopPole": "odpt.BusstopPole:Toei.ShibuyaStation.636.6",
    "fromBusstopPoleTime": "2025-12-01T17:49:13+09:00",
    "startingBusstopPole": "odpt.BusstopPole:Toei.ShibuyaStation.636.6",
    "terminalBusstopPole": "odpt.BusstopPole:Toei.RoppongiHills.2480.1"
  }
]
```

### GET /busstoppole

バス停情報を取得します。

#### パラメータ

- `operator` (必須): 事業者のID（例: `odpt.Operator:Toei`）

#### リクエスト例

```bash
curl "http://localhost:8081/busstoppole?operator=odpt.Operator:Toei"
```

#### レスポンス例

```json
[
  {
    "id": "urn:ucode:_00001C0000000000000100000330CA15",
    "type": "odpt:BusstopPole",
    "sameAs": "odpt.BusstopPole:Toei.Yakuojimachi.1547.1",
    "date": "2025-12-01T03:09:30+09:00",
    "title": "薬王寺町",
    "long": 139.72509,
    "lat": 35.696049,
    "operator": ["odpt.Operator:Toei"]
  }
]
```

#### データソース

バス停情報は以下のローカルJSONファイルから取得されます：

- `assets/odpt_BusstopPole_Toei.json` - 都営バスのバス停情報
- `assets/odpt_BusstopPole_<operator>.json` - その他の事業者のバス停情報

ファイル名は `odpt.Operator:<operator>` の `<operator>` 部分に対応します。

## 元のAPI

このラッパーAPIは以下のODPT APIを使用しています:

- エンドポイント: `https://api-public.odpt.org/api/v4/odpt:Bus`
- パラメータマッピング:
  - ラッパーAPI `operator` → ODPT API `odpt:operator`

## 開発

### ビルド

```bash
go build -o transport-realtime.exe main.go
```

### 実行

```bash
./transport-realtime.exe
```
