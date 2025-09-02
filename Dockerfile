# ビルドステージ
FROM golang:1.25 AS builder

# 作業ディレクトリを設定
WORKDIR /app

# 依存関係をコピーしてダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# アプリケーションをビルド
# CGO_ENABLED=0 は静的リンクを有効にし、distrolessと相性が良い
RUN CGO_ENABLED=0 go build -o /goapp main.go

# 実行ステージ
# gcr.io/distroless/static は、静的リンクされたバイナリの実行に最適
FROM gcr.io/distroless/static-debian12

# タイムゾーン情報をコピー
# Goでタイムゾーンが必要な場合はこの行を追加
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# 作業ディレクトリを設定
WORKDIR /

# ビルドステージからバイナリをコピー
COPY --from=builder /goapp /goapp

# アプリケーションを実行
CMD ["/goapp"]