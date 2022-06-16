# Golang Poker
PokerをGo言語を使って実装しました。現在開発途中です。
テストプレイ用([http://15.152.231.230:8080/] EC2上においています。)
Git Hubから落としてきてdocker-composeすれば8081番ポートで動きます。  
```console
git clone https://github.com/shoheiKU/golang_poker.git
docker-compose build
docker-compose up -d
docker-compose exec web go run ./cmd/web/.
```
使用後はコンテナを閉じてください。
```console
docker-compose down
```
- 使用言語  
	- Golang, JS(jQuery), html, css  
- 使用技術  
	- Docker  
- 実装済みの機能  
	- ランダムなカード配布  
	- 基本的なベット機能  
	- 他プレイヤーのベット情報の取得(Ajax使用)  
	- ベットフェーズの切り替わり情報の取得(Ajax使用)  
	- 自動の勝敗判断システム  
	- コミュニティカードの表示  
	- 連続ゲームの対応  
- 未実装の機能  
	- テスト実装  
	- ユーザー登録  
	- 複数ゲームの実装  
	- サイドポットの対応  
	- etc  

## Home
プレイの方法を書いてます。
## About
コンセプトなどを書いてます。
## Poker
メインページです。コミュニティカードが表示されます。各ゲームの開始などもこのページでコントロールします。
## Mobile Poker
各プレイヤーが表示するページです。各プレイヤーの情報が表示されます。ベットなどの操作をこのページで行います。
## Remote Poker
離れている場所でゲームをプレイするためのページです。基本的にはPokerページとMobile Pokerページを1ページにまとめたページになります。
## Control
Repositoryをリセットするためのページです。
## Contact
コンタクトページです。
