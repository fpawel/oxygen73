SET APP_DIR=build
SET GOARCH=386
buildmingw32 go build -o %APP_DIR%\oxygen73.exe github.com/fpawel/oxygen73/cmd/oxygen73
buildmingw32 go build -o %APP_DIR%\run.exe github.com/fpawel/oxygen73/cmd/run