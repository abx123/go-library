$Env:GOOS = "linux"
$Env:CGO_ENABLED = "0"
$Env:GOARCH = "amd64"

go build -o output/books ./books/main.go
go build -o output/library ./library/main.go

~\Go\Bin\build-lambda-zip.exe -output output/books.zip output/books
~\Go\Bin\build-lambda-zip.exe -output output/library.zip output/library