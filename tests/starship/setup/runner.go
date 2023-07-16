package setup

import (
	"fmt"
	starship "github.com/cosmology-tech/starship/clients/go/client"
	"os"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// CmdRunner will help you to run admin commands on
type CmdRunner struct {
	logger *zap.Logger
	config *starship.Config
}

func NewCmdRunner(logger *zap.Logger, config *starship.Config) (*CmdRunner, error) {
	return &CmdRunner{
		logger: logger,
		config: config,
	}, nil
}

func (c *CmdRunner) runKubectlCmd(cmdStr string) (string, string, error) {
	cmdSplit := strings.Split(cmdStr, " ")
	cmd := exec.Command(cmdSplit[0], cmdSplit[1:]...)
	cmd.Env = os.Environ()
	buf, stderr := new(strings.Builder), new(strings.Builder)
	cmd.Stdout = buf
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		c.logger.Error("unable to run command",
			zap.String("command", cmdStr),
			zap.String("stderr", stderr.String()),
			zap.String("stdout", buf.String()),
			zap.Error(err))
		return "", "", fmt.Errorf("unable to run command: %s, stderr: %s", cmdStr, stderr.String())
	}
	return buf.String(), stderr.String(), nil
}

// GetPodFromName will return the full pod name from a given name
// by querying the kubernetes cluster for the given name
func (c *CmdRunner) GetPodFromName(name string) (string, error) {
	cmd := fmt.Sprintf("kubectl get pods --no-headers -lapp.kubernetes.io/rawname=%s", name)
	out, _, err := c.runKubectlCmd(cmd)
	if err != nil {
		return "", err
	}
	return strings.Split(out, " ")[0], nil
}

// RunExec runs an exec command directly inside the container
// we specify. One can either provide the full container name, or just the
// `name` as defined in the config file.
func (c *CmdRunner) RunExec(name string, cmd string) (string, error) {
	podName, err := c.GetPodFromName(name)
	if err != nil {
		return "", err
	}
	cmdStr := fmt.Sprintf("kubectl exec -it %s -- %s", podName, cmd)
	stdout, stderr, err := c.runKubectlCmd(cmdStr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("stdout: %s, stderr: %s", stdout, stderr), nil
}
