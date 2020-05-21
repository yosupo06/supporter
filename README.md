# Supporter

## 準備

### インストール

- `oj`をインストールする(略)
- goをインストールする
- `go get -v github.com/yosupo06/supporter && go install github.com/yosupo06/supporter`でインストール, `supporter --help`でそれっぽいのが出てくれば成功

- `config_template.toml`を好きな場所にコピーして、内容をいじる
- コピーした`config_template.toml`のパスを`$SUPPORTER_CONFIG`に追加する
  - 例: `.profile`とか`.bash_profile`に`export SUPPORTER_CONFIG="/absuolute/path/of/config_template.toml"`



## 使い方

### Init Command

```
supporter i https://atcoder.jp/contests/agc043 {A..F}
supporter i https://codeforces.com/contest/1349 {A..F}
```

とか

`agc043/A/main.cpp`がいじるファイル

### Build Command

```
supporter b agc043/A    # debug mode
supporter b agc043/A -O # opt mode
cd agc043 && supporter b A
```

### Test Command

```
supporter t agc043/A
```

- `oj`でサンプルをダウンロードして、それのテストをする
- `agc043/A/ourtest`に`hoge.in`, (+ `hoge.out`)とか適当な名前でファイルを置いておくとそれもテストされる

### Submit Command

```
supporter s agc043/A
supporter s agc043/A -c #クリップボードに提出コードをコピー, macでしか動かない気がする
```

- ただの`oj`のラッパー、(`config.toml`で設定していればbundleコマンドを走らせ、)ソースを提出する。
