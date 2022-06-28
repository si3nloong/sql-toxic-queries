package main

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	_ "github.com/go-sql-driver/mysql"
	"github.com/segmentio/ksuid"
	"github.com/si3nloong/sqlike/sql/expr"
	sqlstmt "github.com/si3nloong/sqlike/sql/stmt"
	"github.com/si3nloong/sqlike/sqlike"
	"github.com/si3nloong/sqlike/sqlike/actions"
	"github.com/si3nloong/sqlike/sqlike/options"
)

type Logger struct {
}

func (l Logger) Debug(stmt *sqlstmt.Statement) {
	// log.Printf("%v", stmt)
	log.Printf("%+v", stmt)
}

const tableUser = "User"

type User struct {
	ID      ksuid.KSUID
	Name    string
	Email   string
	Created time.Time
}

func newUser() *User {
	usr := &User{}
	usr.ID = ksuid.New()
	usr.Name = gofakeit.Name() // Markus Moen
	usr.Email = gofakeit.Email()
	usr.Created = time.Now().UTC()
	return usr
}

type Result struct {
	Rows uint64 `sqlike:"rows"`
}

func setup() (*sqlike.Client, *sqlike.Database) {
	ctx := context.Background()
	client := sqlike.MustConnect(
		ctx,
		"mysql",
		options.Connect().
			ApplyURI(`root:abcd1234@tcp()/toxicquery?parseTime=true&loc=UTC&charset=utf8mb4&collation=utf8mb4_general_ci`),
	)

	client.SetPrimaryKey("ID")
	// client.SetLogger(&Logger{})
	db := client.Database("toxicquery")

	table := db.Table(tableUser)
	table.MustUnsafeMigrate(ctx, User{})

	rows, err := db.QueryStmt(ctx, expr.Raw("EXPLAIN SELECT * FROM `User`"))
	if err != nil {
		panic(err)
	}

	rows.Next()
	var result Result
	if err := rows.Decode(&result); err != nil {
		panic(err)
	}
	rows.Close()

	if result.Rows >= 5000 {
		return client, db
	}

	// set timezone for UTC
	for i := 0; i < 10; i++ {
		data := []*User{}
		for j := 0; j < 200; j++ {
			data = append(data, newUser())
		}

		if _, err := table.Insert(ctx, &data); err != nil {
			panic(err)
		}

		time.Sleep(time.Second * 1)
	}

	return client, db
}

func BenchmarkCountStatement(b *testing.B) {
	ctx := context.Background()
	_, db := setup()
	table := db.Table(tableUser)

	b.Run("Count with *", func(b *testing.B) {
		var count uint64
		result, err := table.Find(ctx, actions.Find().
			Select(expr.Count(expr.Raw("*"))))
		if err != nil {
			b.FailNow()
		}
		defer result.Close()

		result.Next()
		if err := result.Scan(&count); err != nil {
			b.FailNow()
		}
	})

	b.Run("Count with Primary Key", func(b *testing.B) {
		var count uint64
		result, err := table.Find(ctx, actions.Find().
			Select(expr.Count("ID")))
		if err != nil {
			b.FailNow()
		}
		defer result.Close()

		result.Next()
		if err := result.Scan(&count); err != nil {
			b.FailNow()
		}
	})

	b.Run("Count with Explain", func(b *testing.B) {
		rows, err := db.QueryStmt(ctx, expr.Raw("EXPLAIN SELECT `ID` FROM `User`;"))
		if err != nil {
			b.FailNow()
		}
		defer rows.Close()

		rows.Next()
		var result Result
		if err := rows.Decode(&result); err != nil {
			b.FailNow()
		}
	})
}

func BenchmarkLikeStatement(b *testing.B) {
	ctx := context.Background()
	limit := uint(100)
	_, db := setup()
	table := db.Table(tableUser)

	b.Run("Like with Leading Wildcard", func(b *testing.B) {
		users := []*User{}
		result, err := table.Find(ctx, actions.Find().
			Where(
				expr.Raw("`Name` LIKE \"%Sibyl Gaylord%\""),
			).Limit(limit))
		if err != nil {
			b.FailNow()
		}
		defer result.Close()

		if err := result.All(&users); err != nil {
			b.FailNow()
		}
	})

	b.Run("Like without Leading Wildcard", func(b *testing.B) {
		users := []*User{}
		result, err := table.Find(ctx, actions.Find().
			Where(
				expr.Like("Name", "Sibyl Gaylord%"),
			).Limit(limit))
		if err != nil {
			b.FailNow()
		}
		defer result.Close()

		if err := result.All(&users); err != nil {
			b.FailNow()
		}
	})
}

func BenchmarkPagination(b *testing.B) {
	ctx := context.Background()
	limit := uint(100)
	_, db := setup()

	b.Run("Offset Based Pagination", func(b *testing.B) {
		offset := uint(0)

		for {
			result, err := db.Table(tableUser).
				Find(ctx,
					actions.Find().Offset(offset*limit).Limit(limit),
					options.Find())
			if err != nil {
				b.FailNow()
			}

			users := []User{}
			if err := result.All(&users); err != nil {
				b.FailNow()
			}

			result.Close()

			noOfRecord := uint(len(users))
			if noOfRecord == 0 {
				break
			}

			if noOfRecord < limit {
				break
			}

			offset++
		}
	})

	b.Run("Cursor Based Pagination", func(b *testing.B) {
		var nextCursor string

		for {
			result, err := db.Table(tableUser).
				Paginate(ctx, actions.Paginate().Limit(limit),
					options.Paginate())
			if err != nil {
				b.FailNow()
			}

			if nextCursor != "" {
				if err := result.NextCursor(ctx, nextCursor); err != nil {
					b.FailNow()
				}
			}

			users := []User{}
			if err := result.All(&users); err != nil {
				b.FailNow()
			}

			noOfRecord := uint(len(users))
			if noOfRecord == 0 {
				break
			}

			if noOfRecord < limit {
				break
			}

			nextCursor = users[noOfRecord-1].ID.String()
		}
	})
}

// func BenchmarkStoreProcedure(b *testing.B) {
// 	setup()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {

// 	}
// }

func main() {
	// var cancel context.CancelFunc
	// ctx, cancel = context.WithCancel(context.Background())
	// defer cancel()

}
