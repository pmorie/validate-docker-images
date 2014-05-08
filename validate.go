package vdc

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/fsouza/go-dockerclient"
)

type AllowedHttpResponses []int

func (r AllowedHttpResponses) Contains(code int) bool {
	candidates := ([]int)(r)

	for i := range candidates {
		if candidates[i] == code {
			return true
		}
	}

	return false
}

type ValidateHttpRequest struct {
	DockerSocket string
	Verbose      bool

	ContainerID string
	Port        string
	Path        string
	Responses   AllowedHttpResponses
	Title       string
}

type ValidateResult struct {
	Valid    bool
	Messages []string
}

var htmlTitleExp = regexp.MustCompile(`<title>([^<]+)</title>`)

func ValidateHttp(req ValidateHttpRequest) (*ValidateResult, error) {
	if req.Port == "" {
		return nil, errors.New("port must be provided")
	}
	if len(req.Responses) == 0 {
		return nil, errors.New("allowed http responses must be provided")
	}

	if req.Verbose {
		log.Printf("Validating container %s for http:\n", req.ContainerID)
		log.Printf("Port: %s\n", req.Port)
		log.Printf("Path: %s\n", req.Path)
		log.Printf("Title: %s\n", req.Title)
		log.Printf("Allowed responses: %+v\n", req.Responses)
	}

	result := &ValidateResult{Valid: true}

	dockerClient, err := docker.NewClient(req.DockerSocket)
	if err != nil {
		return nil, err
	}

	container, err := dockerClient.InspectContainer(req.ContainerID)
	if err != nil {
		return nil, err
	}

	mappedPort := container.NetworkSettings.Ports[docker.Port(req.Port)]
	if req.Verbose {
		log.Printf("Container: %+v\n", container)
		log.Printf("NetworkSettings: %+v\n", container.NetworkSettings)
	}
	log.Printf("Container has port %s mapped to %s:%s\n", req.Port, mappedPort[0].HostIp, mappedPort[0].HostPort)

	url := "http://" + mappedPort[0].HostIp + ":" + mappedPort[0].HostPort
	if req.Path != "" {
		if strings.HasPrefix(req.Path, "/") {
			url += req.Path
		} else {
			url += "/" + req.Path
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if req.Verbose {
		log.Printf("Http response status code: %s\n", resp.Status)
	}

	if !req.Responses.Contains(resp.StatusCode) {
		result.Valid = false
		message := fmt.Sprintf("Invalid response status code: %d\n", resp.StatusCode)
		result.Messages = append(result.Messages, message)
	}

	if req.Title != "" {
		matches := htmlTitleExp.FindAllStringSubmatch(string(body), -1)
		if len(matches) == 0 {
			if req.Verbose {
				log.Println("Response did not contain a title")
			}
		}

		responseTitle := matches[0][1]
		if responseTitle != req.Title {
			result.Valid = false
			result.Messages = append(result.Messages, "Title did not match")
		}

	}

	return result, nil
}

func ValidateHttps(req ValidateHttpRequest) (*ValidateResult, error) {
	return nil, nil
}
