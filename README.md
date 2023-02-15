# このリポジトリについて

Youtube で自動生成された英語字幕データを取得して、Nvidia/Nemoの句読点復元モデルと Google Translation API を用いてより高精度な日本語字幕を生成します

# Install

1. Python と Go の実行環境を用意する

2. Nvidia/Nemoをインストール

3. GCP でプロジェクトを作成して YouTube Data API を登録する

4. 作成した GCP プロジェクト内で API キーを生成する

5. ルートディレクトリに`.env`ファイルを生成して、生成した API キーを貼り付ける
   `YOUTUBE_DATA_API_KEY="XXXXXXXXXXXXXXXXXXXXXXXXXXXXX"`

# 使用方法

和訳したい Youtube 動画の ID を取得します

ID は例えばhttps://www.youtube.com/watch?v=446E-r0rXHI の場合、クエリパラメータの v が ID に相当します

取得した ID を引数に go-youtube-caps-translater を以下のように実行します

`./go-youtube-caps-translater 446E-r0rXHI`

すると、`./captions/446E-r0rXHI`ディレクトリ内に翻訳後の字幕データである`captions_ja.srt`ファイルが生成されます。

生成した字幕ファイルを[ブラウザ拡張機能](https://chrome.google.com/webstore/detail/substital-add-subtitles-t/kkkbiiikppgjdiebcabomlbidfodipjg)等で読み込むことで Youtube 上で表示できます
