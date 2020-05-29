# dotlive-schedule-server

[.スケジュール](https://apps.apple.com/jp/app/%E3%82%B9%E3%82%B1%E3%82%B8%E3%83%A5%E3%83%BC%E3%83%AB/id1512712289?mt=8)のサーバー  
https://dotlive-schedule.appspot.com/

クライアント:[yaegaki/dotlive-schedule-client](https://github.com/yaegaki/dotlive-schedule-client)

## デプロイ方法

app.yamlと同じディレクトリにsecret.yamlを作成し、以下の内容を埋める。

```yaml
env_variables:
  TWITTER_CONSUMER_KEY: "XXXX"
  TWITTER_CONSUMER_SECRET: "XXXX"
  FIREBASE_SERVER_KEY: "XXXX"
```

`TWITTER_CONSUMER_KEY`と`TWITTER_CONSUMER_SECRET`はTwitterのKeys and tokensから取得できる。  
`FIREBASE_SERVER_KEY`はプッシュ通知に使用するキーで設定のクラウドメッセージングから取得できる。

`secret.yaml`を用意したら通常通り以下のコマンドでデプロイできる。

```sh
gcloud app deploy
```