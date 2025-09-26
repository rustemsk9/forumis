package data

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"

	"forum/utils"

	_ "github.com/mattn/go-sqlite3"
)

type LoginSkin struct {
	Submit string
	Signup string
	Name   string
	Email  string
	Error  string
}

type sqlInfo struct {
	DriverName string
	Username   string
	Password   string
	Database   string
}

var info sqlInfo
var Db *sql.DB

func init() {
	var err error
	Db, err = sql.Open("sqlite3", "mydb.db")
	if err != nil {
		log.Fatal(err)
	}
}

// create a random UUID with from RFC 4122
// adapted from http://github.com/nu7hatch/gouuid
func createUUID() (uuid string) {
	u := new([16]byte)
	_, err := rand.Read(u[:])
	if err != nil {
		utils.Danger("Cannot generate UUID", err)
	}

	// 0x40 is reserved variant from RFC 4122

	u[8] = (u[8] | 0x40) & 0x7F
	// Set the four most significant bits (bits 12 through 15) of the
	// time_hi_and_version field to the 4-bit version number.
	u[6] = (u[6] & 0xF) | (0x4 << 4)
	uuid = fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
	return
}

// hash plaintext with SHA-1, changed to bcrypt as in forum instructions.
func Encrypt(plaintext string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	return string(hash)
}

func CheckPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
