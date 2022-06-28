package main

import (
	"context"
	"log"
	"strconv"
	"testing"

	"github.com/segmentio/ksuid"
	"github.com/si3nloong/sqlike/sql/expr"
	sqlstmt "github.com/si3nloong/sqlike/sql/stmt"
	"github.com/si3nloong/sqlike/sqlike/actions"
	"github.com/si3nloong/sqlike/sqlike/options"
)

type Logger struct {
}

func (l Logger) Debug(stmt *sqlstmt.Statement) {
	// log.Printf("%v", stmt)
	log.Printf("%+v", stmt)
}

type Result struct {
	Rows uint64 `sqlike:"rows"`
}

func BenchmarkJoinStatement(b *testing.B) {
	ctx := context.Background()
	_, db := setup()
	emails := []string{
		"aaliyahheathcote@douglas.org",
		"heidischiller@casper.info",
		"zorabeatty@wintheiser.name",
	}
	query := `
SELECT * FROM Car AS c LEFT JOIN User AS u
ON c.UserID = u.ID
WHERE u.Email IN (`
	for i, email := range emails {
		if i > 0 {
			query += ","
		}
		query += strconv.Quote(email)
	}
	query += ");"

	b.ResetTimer()
	b.Run("SELECT with JOIN", func(b *testing.B) {
		result, err := db.QueryStmt(ctx, expr.Raw(query))
		if err != nil {
			log.Println(err)
			b.FailNow()
		}
		defer result.Close()

		userCars := []UserCar{}
		for result.Next() {
			var userCar UserCar
			if err := result.Decode(&userCar); err != nil {
				log.Println(err)
				b.FailNow()
			}

			userCars = append(userCars, userCar)
		}
	})

	b.Run("SELECT without JOIN", func(b *testing.B) {
		result, err := db.Table(tableUser).Find(
			ctx,
			actions.Find().
				Where(
					expr.In("Email", emails),
				),
		)
		if err != nil {
			b.FailNow()
		}

		users := make(map[ksuid.KSUID]User)
		userIDs := []ksuid.KSUID{}
		for result.Next() {
			var user User
			if err := result.Decode(&user); err != nil {
				b.FailNow()
			}

			userIDs = append(userIDs, user.ID)
			users[user.ID] = user
		}

		result.Close()

		result, err = db.Table(tableCar).Find(
			ctx,
			actions.Find().
				Where(
					expr.In("UserID", userIDs),
				),
		)
		if err != nil {
			b.FailNow()
		}

		userCars := []UserCar{}
		for result.Next() {
			var car Car
			if err := result.Decode(&car); err != nil {
				b.FailNow()
			}

			userCars = append(userCars, UserCar{
				User: users[car.UserID],
				Car:  car,
			})
		}

		result.Close()

	})
}

func BenchmarkCountStatement(b *testing.B) {
	ctx := context.Background()
	_, db := setup()
	table := db.Table(tableUser)
	b.ResetTimer()

	b.Run("COUNT with *", func(b *testing.B) {
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

	b.Run("COUNT with Primary Key", func(b *testing.B) {
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

	b.Run("COUNT with Explain", func(b *testing.B) {
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
	b.ResetTimer()

	b.Run("LIKE with Leading Wildcard", func(b *testing.B) {
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

	b.Run("LIKE without Leading Wildcard", func(b *testing.B) {
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
	table := db.Table(tableUser)
	b.ResetTimer()

	b.Run("Offset Based Pagination", func(b *testing.B) {
		offset := uint(0)

		for {
			result, err := table.
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
			result, err := table.
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

func BenchmarkStoreProcedure(b *testing.B) {
	ctx := context.Background()
	client, db := setup()
	b.ResetTimer()

	b.Run("INSERT with Stored Procedure", func(b *testing.B) {
		// DELIMITER $$
		// CREATE PROCEDURE insertUser(IN id VARCHAR(255), IN name VARCHAR(255), IN email VARCHAR(255), IN created VARCHAR(255))
		// BEGIN
		// 	INSERT INTO User (ID, Name, Email, Created) VALUES (id, name, email, created);
		// END $$
		// DELIMITER ;

		b.StopTimer()
		user := newUser()
		b.StartTimer()

		if _, err := client.ExecContext(ctx, `call insertUser(?,?,?,?)`,
			user.ID.String(),
			user.Name,
			user.Email,
			user.Created,
		); err != nil {
			b.FailNow()
		}
	})

	b.Run("INSERT", func(b *testing.B) {
		b.StopTimer()
		user := newUser()
		b.StartTimer()

		if _, err := db.Table(tableUser).
			InsertOne(ctx, user); err != nil {
			b.FailNow()
		}
	})
}
