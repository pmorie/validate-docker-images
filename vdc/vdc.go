package main

import (
	"log"
	_ "net/http/pprof"
	"strconv"
	"strings"

	vdc "github.com/pmorie/validate-docker-images"
	"github.com/spf13/cobra"
)

func parseValidCodes(input string) ([]int, error) {
	var intPorts []int
	ports := strings.Split(input, ",")
	for i := range ports {
		intPort, err := strconv.Atoi(ports[i])
		if err != nil {
			return nil, err
		}

		intPorts = append(intPorts, intPort)
	}

	return intPorts, nil
}

func validateHttpArgs(httpReq vdc.ValidateHttpRequest, responseCodes string) bool {
	ok := true

	if httpReq.Port == "" {
		log.Println("You must specify a port to check")
		ok = false
	}

	if responseCodes == "" {
		log.Println("You must specify valid http response codes")
		ok = false
	}

	return ok
}

func Execute() {
	var (
		responseCodes string
		httpReq       vdc.ValidateHttpRequest
	)

	valCmd := &cobra.Command{
		Use:   "vdc",
		Short: "Validate a docker container",
		Long:  "Validate a docker container",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
	valCmd.PersistentFlags().StringVarP(&(httpReq.DockerSocket), "url", "U", "unix:///var/run/docker.sock", "Set the url of the docker socket to use")
	valCmd.PersistentFlags().BoolVar(&(httpReq.Verbose), "verbose", false, "Enable verbose output")
	valCmd.PersistentFlags().StringVarP(&(httpReq.Port), "port", "p", "", "Set the port to check")
	valCmd.PersistentFlags().StringVar(&(httpReq.Path), "P", "", "Specify a path to validate with an HTTP request")
	valCmd.PersistentFlags().StringVarP(&responseCodes, "responseCodes", "c", "", "A comma-delimited list of response codes")
	valCmd.PersistentFlags().StringVarP(&(httpReq.Title), "title", "t", "", "Specify an HTML title to validate against")

	tcpCmd := &cobra.Command{
		Use:   "tcp CONTAINER_ID PORT",
		Short: "Test connectivity to a container",
		Long:  "Test connectivity to a container",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	valCmd.AddCommand(tcpCmd)

	httpCmd := &cobra.Command{
		Use:   "http <container id>",
		Short: "Test http connectivity to a container",
		Long:  "Test http connectivity to a container",
		Run: func(cmd *cobra.Command, args []string) {
			if !validateHttpArgs(httpReq, responseCodes) {
				return
			}

			httpReq.ContainerID = args[0]
			if !strings.HasSuffix(httpReq.Port, "/tcp") {
				httpReq.Port += "/tcp"
			}
			codes, err := parseValidCodes(responseCodes)
			if err != nil {
				log.Printf("Error parsing response codes: %s\n", err.Error())
				return
			}
			httpReq.Responses = vdc.AllowedHttpResponses(codes)

			res, err := vdc.ValidateHttp(httpReq)
			if err != nil {
				log.Printf("%s\n", err.Error())
				return
			}

			if !res.Valid {
				log.Println("Container failed validation:")
			} else {
				log.Println("Container passed validation:")
			}

			for _, msg := range res.Messages {
				log.Println(msg)
			}
		},
	}
	valCmd.AddCommand(httpCmd)

	httpsCmd := &cobra.Command{
		Use:   "https <container id>",
		Short: "Test https connectivity to a container",
		Long:  "Test https connectivity to a container",
		Run: func(cmd *cobra.Command, args []string) {
			if !validateHttpArgs(httpReq, responseCodes) {
				return
			}

			httpReq.ContainerID = args[0]
			if !strings.HasSuffix(httpReq.Port, "/tcp") {
				httpReq.Port += "/tcp"
			}
			codes, err := parseValidCodes(responseCodes)
			if err != nil {
				log.Printf("Error parsing response codes: %s\n", err.Error())
				return
			}
			httpReq.Responses = vdc.AllowedHttpResponses(codes)

			res, err := vdc.ValidateHttps(httpReq)
			if err != nil {
				log.Printf("%s\n", err.Error())
				return
			}

			if !res.Valid {
				log.Println("Container failed validation:")
			} else {
				log.Println("Container passed validation:")
			}

			for _, msg := range res.Messages {
				log.Println(msg)
			}
		},
	}
	valCmd.AddCommand(httpsCmd)
	valCmd.Execute()
}

func main() {
	Execute()
}
