package pkg

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func populate() {
	f, err := os.Open("load.sql")
	if err != nil {
		panic(err)
	}

	b := make([]byte, 1024)
	n, err := f.Read(b)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b[:n]))

	db, err := sql.Open("mysql", "test:mysql@tcp(localhost:3306)/mysql?multiStatements=true")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	sql := string(b[:n])
	_, err = db.Exec(sql)
	if err != nil {
		panic(err)
	}
	result, err := db.Query("select name from friends")
	if err != nil {
		panic(err)
	}
	for result.Next() {
		colonne := ""
		err = result.Scan(&colonne)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(colonne)
	}
}
