package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"

	"github.com/facebookincubator/ttpforge/pkg/logging"
	"github.com/spf13/cobra"
)

type ExecutionResult struct {
	ID                string `json:"id"`
	Output            string `json:"output"`
	Stderr            string `json:"stderr"`
	ExitCode          string `json:"exit_code"`
	Status            string `json:"status"`
	PID               string `json:"pid"`
	AgentReportedTime string `json:"agent_reported_time"`
}

type BeaconRequest struct {
	Paw               string            `json:"paw"`
	Server            string            `json:"server"`
	Group             string            `json:"group"`
	Host              string            `json:"host"`
	Contact           string            `json:"contact"`
	Username          string            `json:"username"`
	Architecture      string            `json:"architecture"`
	Platform          string            `json:"platform"`
	Location          string            `json:"location"`
	PID               int               `json:"pid"`
	PPID              int               `json:"ppid"`
	Executors         []string          `json:"executors"`
	Privilege         string            `json:"privilege"`
	ExeName           string            `json:"exe_name"`
	ProxyReceivers    string            `json:"proxy_receivers"`
	OriginLinkID      string            `json:"origin_link_id"`
	DeadmanEnabled    bool              `json:"deadman_enabled"`
	AvailableContacts []string          `json:"available_contacts"`
	HostIPAddrs       []string          `json:"host_ip_addrs"`
	UpstreamDest      string            `json:"upstream_dest"`
	Results           []ExecutionResult `json:"results"`
}

type Instruction struct {
	Deadman       bool     `json:"deadman"`
	ID            string   `json:"id"`
	Sleep         int      `json:"sleep"`
	Command       string   `json:"command"`
	Executor      string   `json:"executor"`
	Timeout       float32  `json:"timeout"`
	Payloads      []string `json:"payloads"`
	DeletePayload bool     `json:"delete_payload"`
	Uploads       []string `json:"uploads"`
}

type BeaconResponse struct {
	Instructions   string  `json:"instructions"`
	Sleep          int     `json:"sleep"`
	Watchdog       int     `json:"watchdog"`
	Paw            string  `json:"paw"`
	NewContact     *string `json:"new_contact";omitempty`
	ExecutorChange *string `json:"executor_change";omitempty`
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	// TODO: FS traversal protection
	logging.L().Infof("File download request received %s", r.Header.Get("file"))
	// Get the file path from the URL
	filePath := r.Header.Get("file")
	if filePath == "" {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}
	// Check if the file exists
	cwd, err := os.Getwd()
	if err != nil {
		http.Error(w, "Failed to get current working directory", http.StatusInternalServerError)
		return
	}
	fullFilePath := cwd + "/" + filePath
	if _, err := os.Stat(fullFilePath); err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Filename", filePath)

	// Serve the file
	http.ServeFile(w, r, filePath)
}

func keepFile(w http.ResponseWriter, r *http.Request) {
	// TODO: Write to arbitrary file
	logging.L().Infof("File upload request received %s", r.Header.Get("X-Request-Id"))
	// Get the file from the request
	file, fh, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	agentID := r.Header.Get("X-Request-Id")
	cwd, err := os.Getwd()
	if err != nil {
		http.Error(w, "Failed to get current working directory", http.StatusInternalServerError)
		return
	}

	// Create a new file on disk
	dstDir := fmt.Sprintf("%s/uploads/%s", cwd, agentID)
	err = os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dst := fmt.Sprintf("%s/%s", dstDir, fh.Filename)
	out, err := os.Create(dst)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Copy the uploaded file to the new file on disk
	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return a success
	w.WriteHeader(http.StatusOK)
}

func getInstructionsByAgentID(_ string) []Instruction {
	result := []Instruction{}
	dice := rand.Intn(6) + 1
	logging.L().Infof("Dice roll is %d", dice)
	if dice == 6 {
		deadman := Instruction{
			Deadman:  true,
			ID:       "deadman",
			Sleep:    0,
			Command:  "wall flynn was here",
			Executor: "sh",
			Timeout:  1,
			Payloads: []string{},
		}
		seppuku := Instruction{
			Deadman:       false,
			ID:            "banzai",
			Sleep:         0,
			Command:       "echo Banzai!",
			Executor:      "sh",
			Timeout:       1,
			Payloads:      []string{},
			DeletePayload: false,
			Uploads:       []string{},
		}
		result = append(result, []Instruction{
			seppuku,
			deadman,
		}...)
	} else if dice <= 3 {
		date := Instruction{
			Deadman:  false,
			ID:       "date",
			Sleep:    0,
			Command:  "date",
			Executor: "sh",
			Timeout:  2,
			Payloads: []string{},
		}
		ttpforge := Instruction{
			Deadman:       false,
			ID:            "ttpforge",
			Sleep:         60,
			Command:       "./ttpforge run --help",
			Executor:      "sh",
			Timeout:       2,
			Payloads:      []string{"ttpforge"},
			DeletePayload: true,
		}
		result = append(result, []Instruction{
			date,
			ttpforge,
		}...)
	} else if dice > 3 {
		exfil := Instruction{
			Deadman:  false,
			ID:       "exfil",
			Sleep:    0,
			Command:  "echo steal it now",
			Executor: "sh",
			Timeout:  1,
			Payloads: []string{},
			Uploads:  []string{"/home/nesusvet/exfil.txt"},
		}
		result = append(result, exfil)
	}

	return result
}

func beaconHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	logging.L().Infof("Beacon request received from %s", r.RemoteAddr)

	// Decode the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	beaconRequest, err := parseRequestBody(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Report execution results if any
	if beaconRequest.Results != nil {
		reportResults(beaconRequest.Results)
	}

	instructions := getInstructionsByAgentID(beaconRequest.Paw)
	textInstructions, err := serializeInstructions(instructions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	beaconResponse := BeaconResponse{
		Instructions:   string(textInstructions),
		Sleep:          60,
		Watchdog:       300,
		Paw:            beaconRequest.Paw,
		NewContact:     nil,
		ExecutorChange: nil,
	}

	responseBytes, err := serializeResponse(beaconResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseBytes))
}

func parseRequestBody(body string) (*BeaconRequest, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return nil, err
	}

	beaconRequest := BeaconRequest{}
	err = json.Unmarshal(decodedBytes, &beaconRequest)
	if err != nil {
		return nil, err
	}
	return &beaconRequest, nil
}

func reportResults(results []ExecutionResult) {
	for _, res := range results {
		logging.L().Infof("Got exec results for %v", res.ID)
		encodedOutput := res.Output
		decodedOutput, err := base64.StdEncoding.DecodeString(encodedOutput)
		if err != nil {
			logging.L().Errorf("Failed to decode output: %v", err)
		} else {
			logging.L().Infof("Output %v", string(decodedOutput))
		}
		encodedStderr := res.Stderr
		decodedStderr, err := base64.StdEncoding.DecodeString(encodedStderr)
		if err != nil {
			logging.L().Errorf("Failed to decode stderr: %v", err)
		} else {
			logging.L().Infof("Stderr %v", string(decodedStderr))
		}
	}
}

func serializeInstructions(instructions []Instruction) (string, error) {
	jsonInstructions := []string{}
	for _, inst := range instructions {
		encodedCommand := base64.StdEncoding.EncodeToString([]byte(inst.Command))
		copyInstruction := Instruction{
			Deadman:       inst.Deadman,
			ID:            inst.ID,
			Sleep:         inst.Sleep,
			Command:       encodedCommand, // Change only this
			Executor:      inst.Executor,
			Timeout:       inst.Timeout,
			Payloads:      inst.Payloads,
			DeletePayload: inst.DeletePayload,
			Uploads:       inst.Uploads,
		}
		byteInstructions, err := json.Marshal(copyInstruction)
		if err != nil {
			return "", err
		}
		jsonInstructions = append(jsonInstructions, string(byteInstructions))
	}
	resultBytes, err := json.Marshal(jsonInstructions)
	if err != nil {
		return "", err
	}
	return string(resultBytes), nil
}

func serializeResponse(response BeaconResponse) (string, error) {
	jsonBytes, err := json.Marshal(response)
	logging.L().Debugf("Built response JSON %v", string(jsonBytes))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(jsonBytes), nil
}

func serve() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world!")
	})
	http.HandleFunc("/file/download", serveFile)
	http.HandleFunc("/file/upload", keepFile)
	http.HandleFunc("/beacon", beaconHandler)

	logging.L().Info("Starting HTTP server now")
	return http.ListenAndServe(":8888", nil)
}

func buildServeCommand(_ *Config) *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "serve --port 8080",
		Short: "Run the C&C server for Mitre sandcat",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// don't want confusing usage display for errors past this point
			cmd.SilenceUsage = true

			err := serve()
			if err != nil {
				return fmt.Errorf("failed to run server")
			}
			return nil
		},
	}
	// runCmd.Flags().StringToIntVarP(&args, "port", "p", 8080, "TCP port to listen on")

	return runCmd
}
