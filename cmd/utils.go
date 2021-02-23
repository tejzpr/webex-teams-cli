package cmd

import (
	"crypto/md5"
	b64 "encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"hash/adler32"
	"io"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

func (app *Application) isValidUrl(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	return true
}

func (app *Application) getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (app *Application) getFilenameWithoutExtension(filePath string) (string, string) {
	return strings.TrimSuffix(filePath, path.Ext(filePath)), path.Ext(filePath)
}

func (app *Application) validateUUID(str string) error {
	err := validation.Validate(str, validation.Required, is.UUID)
	if err != nil {
		return err
	}
	return nil
}

func (app *Application) getAdlerHash(str string) string {
	adlerHash := adler32.New()
	adlerHash.Write([]byte(str))
	return fmt.Sprint(adlerHash.Sum32())
}

func (app *Application) parseRoomID(str string) (string, error) {
	err := app.validateUUID(str)
	if err == nil {
		return str, nil
	}

	sDec, err := b64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	urn := string(sDec)
	parts := strings.Split(urn, "/ROOM/")
	err = app.validateUUID(parts[1])
	if err != nil {
		return "", err
	}
	return parts[1], nil
}

// userCSV struct
type userCSV struct {
	Email       email `csv:"email"`
	IsModerator bool  `csv:"moderator,omitempty"`
}

// UserCSVReturn struct
type UserCSVReturn struct {
	Value userCSV
	Err   error
}

// ParseUsersCSV parse a csv file and return an array of resources
func ParseUsersCSV(r io.Reader) chan UserCSVReturn {
	c := make(chan UserCSVReturn, 10)
	go func() {
		defer close(c)
		rd := csv.NewReader(r)
		var header []string
		header, err := rd.Read()
		if err != nil {
			c <- UserCSVReturn{userCSV{}, err}
		}

		e := userCSV{}
		et := reflect.TypeOf(e)
		var headers = make(map[string]int, et.NumField())
		for i := 0; i < et.NumField(); i++ {
			headers[et.Field(i).Name] = func(element string, array []string) int {
				for k, v := range array {
					if v == element {
						return k
					}
				}
				return -1
			}(et.Field(i).Tag.Get("csv"), header)
		}
		for {
			var e = userCSV{}
			record, err := rd.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				c <- UserCSVReturn{userCSV{}, err}
			}
			for h, i := range headers {
				if i == -1 {
					continue
				}
				elem := reflect.ValueOf(&e).Elem()
				field := elem.FieldByName(h)
				if field.CanSet() {
					switch field.Type().Name() {
					case "float64":
						a, _ := strconv.ParseFloat(record[i], 64)
						field.Set(reflect.ValueOf(a))
					case "email":
						validationErr := validation.Validate(record[i], validation.Required, is.Email)
						if validationErr == nil {
							field.Set(reflect.ValueOf(email(record[i])))
						} else {
							c <- UserCSVReturn{userCSV{}, fmt.Errorf("%s is not a valid email", record[i])}
						}
					case "Time":
						a, _ := time.Parse("2006-01-02T00:00:00Z", record[i])
						field.Set(reflect.ValueOf(a))
					case "bool":
						a, err := strconv.ParseBool(record[i])
						if err != nil {
							c <- UserCSVReturn{userCSV{}, fmt.Errorf("%s is not a valid moderator flag. Correct values are either true/false", record[i])}
						}
						field.Set(reflect.ValueOf(a))
					default:
						field.Set(reflect.ValueOf(record[i]))
					}
				}
			}
			c <- UserCSVReturn{e, nil}
		}
	}()
	return c
}

// roomsCSV struct
type roomsCSV struct {
	RoomID string `csv:"roomids"`
}

// RoomsIDsCSVReturn struct
type RoomsIDsCSVReturn struct {
	Value roomsCSV
	Err   error
}

// ParseRoomIDsCSV parse a csv file and return an array of resources
func ParseRoomIDsCSV(r io.Reader) chan RoomsIDsCSVReturn {
	c := make(chan RoomsIDsCSVReturn, 10)
	go func() {
		defer close(c)
		rd := csv.NewReader(r)
		var header []string
		header, err := rd.Read()
		if err != nil {
			c <- RoomsIDsCSVReturn{roomsCSV{}, err}
		}

		e := roomsCSV{}
		et := reflect.TypeOf(e)
		var headers = make(map[string]int, et.NumField())
		for i := 0; i < et.NumField(); i++ {
			headers[et.Field(i).Name] = func(element string, array []string) int {
				for k, v := range array {
					if v == element {
						return k
					}
				}
				return -1
			}(et.Field(i).Tag.Get("csv"), header)
		}
		for {
			var e = roomsCSV{}
			record, err := rd.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				c <- RoomsIDsCSVReturn{roomsCSV{}, err}
			}
			for h, i := range headers {
				if i == -1 {
					continue
				}
				elem := reflect.ValueOf(&e).Elem()
				field := elem.FieldByName(h)
				if field.CanSet() {
					switch field.Type().Name() {
					case "float64":
						a, _ := strconv.ParseFloat(record[i], 64)
						field.Set(reflect.ValueOf(a))
					case "email":
						validationErr := validation.Validate(record[i], validation.Required, is.Email)
						if validationErr == nil {
							field.Set(reflect.ValueOf(email(record[i])))
						} else {
							c <- RoomsIDsCSVReturn{roomsCSV{}, fmt.Errorf("%s is not a valid email", record[i])}
						}
					case "Time":
						a, _ := time.Parse("2006-01-02T00:00:00Z", record[i])
						field.Set(reflect.ValueOf(a))
					case "bool":
						a, err := strconv.ParseBool(record[i])
						if err != nil {
							c <- RoomsIDsCSVReturn{roomsCSV{}, fmt.Errorf("%s is not a valid moderator flag. Correct values are either true/false", record[i])}
						}
						field.Set(reflect.ValueOf(a))
					default:
						field.Set(reflect.ValueOf(record[i]))
					}
				}
			}
			c <- RoomsIDsCSVReturn{e, nil}
		}
	}()
	return c
}
