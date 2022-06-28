package main

import (
	"context"
	"math/rand"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/ksuid"
	"github.com/si3nloong/sqlike/sqlike"
	"github.com/si3nloong/sqlike/sqlike/options"
)

const (
	dbName = "toxicquery"

	tableUser = "User"
	tableCar  = "Car"
)

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
	db := client.Database(dbName)

	return client, db
}

type User struct {
	ID      ksuid.KSUID
	Name    string
	Email   string
	Created time.Time
}

type Car struct {
	ID        ksuid.KSUID
	Model     string
	FuelType  string
	Maker     string
	UserID    ksuid.KSUID `sqlike:",foreign_key=User:ID"`
	Purchased time.Time
}

type UserCar struct {
	User
	Car
	ID ksuid.KSUID
}

func newUser() *User {
	usr := &User{}
	usr.ID = ksuid.New()
	usr.Name = gofakeit.Name() // Markus Moen
	usr.Email = gofakeit.Email()
	usr.Created = time.Now().UTC()
	return usr
}

func newCar(user *User) *Car {
	car := &Car{}
	car.ID = ksuid.New()
	car.UserID = user.ID
	car.Model = gofakeit.CarModel()
	car.Maker = gofakeit.CarMaker()
	car.FuelType = gofakeit.CarFuelType()
	car.Purchased = gofakeit.Date()
	return car
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client, db := setup()
	defer client.Close()

	db.Table(tableUser).MustUnsafeMigrate(ctx, User{})
	db.Table(tableCar).MustUnsafeMigrate(ctx, Car{})

	// defer cancel()

	for i := 0; i < 100; i++ {
		users := []*User{}
		for j := 0; j < 500; j++ {
			users = append(users, newUser())
		}

		if _, err := db.Table(tableUser).Insert(ctx, &users); err != nil {
			panic(err)
		}

		noOfUser := len(users)
		num := rand.Intn(10)
		cars := []*Car{}
		for j := 0; j < num; j++ {
			user := users[rand.Intn(noOfUser)]
			cars = append(cars, newCar(user))
		}

		if len(cars) > 0 {
			if _, err := db.Table(tableCar).Insert(ctx, &cars); err != nil {
				panic(err)
			}
		}

	}
}
