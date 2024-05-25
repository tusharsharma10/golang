package helpers

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"restapi/logger"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Response struct {
	Data   interface{} `json:"Data,omitempty"`
	Errors []Error     `json:"Errors,omitempty"`
}

func NewResponse(data interface{}, errors []Error) Response {
	toBeSentData := data
	if toBeSentData == nil {
		toBeSentData = struct{}{}
	}

	return Response{
		Data:   toBeSentData,
		Errors: errors,
	}
}

type Error struct {
	Message string `json:"Message,omitempty"`
	Code    int    `json:"Code,omitempty"`
}

func ValidationError(message string) Error {
	return Error{Code: http.StatusBadRequest, Message: message}
}

func InternalServerError(message string) Error {
	return Error{Code: http.StatusInternalServerError, Message: message}
}

func NotFoundError(message string) Error {
	return Error{Code: http.StatusNotFound, Message: message}
}

func NoContentError(message string) Error {
	return Error{Code: http.StatusNoContent, Message: message}
}

func (err Error) Error() string {
	return err.Message
}

func (err Error) Status() int {
	return err.Code
}

func Recover(c *gin.Context, apiPath string) {
	if r := recover(); r != nil {
		err, valid := r.(Error)

		if !valid {
			err = InternalServerError(fmt.Sprintf("%v", r))
		}

		logger.Error(c, err.Message, logger.Z{"errCode": err.Code, "apiPath": apiPath})

		if os.Getenv("GIN_MODE") == "release" {
			switch err.Code {
			case http.StatusNotFound:
				{
					err = NotFoundError("No results/resource found")

					break
				}
			case http.StatusBadRequest:
				{
					err = ValidationError("Bad Request")

					break
				}
			case http.StatusNoContent:
				{
					err = NoContentError("No Content")

					break
				}
			case http.StatusInternalServerError:
				fallthrough
			default:
				{
					err = InternalServerError("Some error occurred. Please try again later.")

					break
				}
			}
		}

		// if we are here, we should override any
		// previous headers set, just in case
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		c.AbortWithStatusJSON(err.Code, err)
	}
}

func InitiateLoggerAndLoadEnv(logFileName string) string {
	logFileName = strings.ToLower(logFileName)
	environment := flag.String("e", "development", "")

	flag.Usage = func() {
		log.Println("Usage: server -e {mode}")
		os.Exit(1)
	}

	flag.Parse()

	envFile := "." + *environment + ".env"

	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	ap := path.Join(basepath, "../config", envFile)

	if err := godotenv.Load(ap); err != nil {
		log.Fatalf("%s", err)
	}

	logger.Init(logFileName, os.Getenv("LOG_LEVEL"))

	return envFile
}
