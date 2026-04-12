go get github.com/go-sql-driver/mysql@v1.8.1
go mod tidy
sed -i 's/^go 1.*/go 1.23/' go.mod
sed -i '/^toolchain /d' go.mod
