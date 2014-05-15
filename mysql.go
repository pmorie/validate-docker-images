package vdc

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/fsouza/go-dockerclient"
	_ "github.com/go-sql-driver/mysql"
)

const (
	MYSQL_PORT string = "3306"
)

type ValidateMysqlRequest struct {
	ValidateRequest

	Username string
	Password string
}

func (req ValidateMysqlRequest) log() {
	log.Printf("Username: %s\n", req.Username)
	log.Printf("Password: %s\n", req.Password)
}

func requestDB(httpReq *ValidateHttpRequest, mysqlReq ValidateMysqlRequest) string {

	if httpReq.Port == "" {
		httpReq.Port = MYSQL_PORT
	}

	mysqlHost := fmt.Sprintf("%s:%s", httpReq.Path, httpReq.Port)
	credentials := fmt.Sprintf("%s:%s@tcp", mysqlReq.Username, mysqlReq.Password)
	url := fmt.Sprintf("%s(%s)/", credentials, mysqlHost)
	return url
}

func ValidateMysql(httpReq ValidateHttpRequest, mysqlReq ValidateMysqlRequest) (*ValidateResult, error) {
	if mysqlReq.Verbose {
		log.Printf("Validating container %s for mysql:\n", mysqlReq.ContainerID)
	}

	return validateMysql(httpReq, mysqlReq)
}

func validateMysql(httpReq ValidateHttpRequest, mysqlReq ValidateMysqlRequest) (*ValidateResult, error) {

	dockerClient, err := docker.NewClient(mysqlReq.DockerSocket)
	container, err := dockerClient.InspectContainer(mysqlReq.ContainerID)
	httpReq.Path = container.NetworkSettings.IPAddress

	url := requestDB(&httpReq, mysqlReq)

	if mysqlReq.Verbose {
		httpReq.log()
		mysqlReq.log()
	}

	result := &ValidateResult{Valid: true}

	db, err := sql.Open("mysql", url)

	defer db.Close()

	if err != nil {
		message := fmt.Sprintf("-> Unable to open connection to MySQL database on host %s:%s", httpReq.Path, httpReq.Port)
		result.Messages = append(result.Messages, message)
		result.Valid = false
		return result, err
	} else {
		message := fmt.Sprintf("-> Able to connect to MySQL database on host %s:%s", httpReq.Path, httpReq.Port)
		result.Messages = append(result.Messages, message)
	}

	err = db.Ping()

	if err != nil {
		message := fmt.Sprintf("-> Unable to verify connection to the MySQL database on host %s:%s", httpReq.Path, httpReq.Port)
		result.Messages = append(result.Messages, message)
		result.Valid = false
		return result, err
	} else {
		message := fmt.Sprintf("-> Able to verify connection to the MySQL database on host %s:%s", httpReq.Path, httpReq.Port)
		result.Messages = append(result.Messages, message)
	}

	_, err = db.Query("SELECT 1")

	if err != nil {
		message := fmt.Sprintf("-> Unable to get proper reply from MySQL database on host %s:%s", httpReq.Path, httpReq.Port)
		result.Messages = append(result.Messages, message)
		result.Valid = false
		return result, err
	} else {
		message := fmt.Sprintf("-> Able to get proper reply from MySQL database on host %s:%s", httpReq.Path, httpReq.Port)
		result.Messages = append(result.Messages, message)
	}

	return result, nil
}
