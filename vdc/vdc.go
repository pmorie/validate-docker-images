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
			httpReq.ContainerID = args[0]

			if httpReq.Port == "" {
				log.Println("You must specify a port to check")
			} else if !strings.HasSuffix(httpReq.Port, "/tcp") {
				httpReq.Port += "/tcp"
			}

			if responseCodes == "" {
				log.Println("You must specify valid http response codes")
				return
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
			}

			log.Printf("Result: %b\n", res.Valid)
		},
	}
	httpCmd.Flags().StringVarP(&(httpReq.Port), "port", "p", "", "Set the port to check")
	httpCmd.Flags().StringVar(&(httpReq.Path), "P", "", "Specify a path to validate with an HTTP request")
	httpCmd.Flags().StringVarP(&responseCodes, "responseCodes", "c", "", "A comma-delimited list of response codes")
	httpCmd.Flags().StringVarP(&(httpReq.Title), "title", "t", "", "Specify an HTML title to validate against")
	valCmd.AddCommand(httpCmd)

	httpsCmd := &cobra.Command{
		Use:   "https <container id> <PORT> <allowed responses>",
		Short: "Test https connectivity to a container",
		Long:  "Test https connectivity to a container",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	valCmd.AddCommand(httpsCmd)
	valCmd.Execute()
}

func main() {
	Execute()
}
