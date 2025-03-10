package getter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/armosec/armoapi-go/armotypes"
	"github.com/kubescape/opa-utils/reporthandling"
	"github.com/kubescape/opa-utils/reporthandling/attacktrack/v1alpha1"
)

// =======================================================================================================================
// ============================================== LoadPolicy =============================================================
// =======================================================================================================================
var DefaultLocalStore = getCacheDir()

func getCacheDir() string {
	defaultDirPath := ".kubescape"
	if homeDir, err := os.UserHomeDir(); err == nil {
		defaultDirPath = filepath.Join(homeDir, defaultDirPath)
	}
	return defaultDirPath
}

// Load policies from a local repository
type LoadPolicy struct {
	filePaths []string
}

func NewLoadPolicy(filePaths []string) *LoadPolicy {
	return &LoadPolicy{
		filePaths: filePaths,
	}
}

// GetControl returns a control from the policy file.
func (lp *LoadPolicy) GetControl(controlID string) (*reporthandling.Control, error) {
	control := &reporthandling.Control{}
	filePath := lp.filePath()

	f, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(f, control); err != nil {
		return control, err
	}

	if controlID == "" || strings.EqualFold(controlID, control.ControlID) {
		return control, nil
	}

	framework, err := lp.GetFramework(control.Name)
	if err != nil {
		return nil, fmt.Errorf("control from file not matching")
	}

	for _, toPin := range framework.Controls {
		ctrl := toPin
		if strings.EqualFold(ctrl.ControlID, controlID) {
			control = &ctrl

			break
		}
	}

	return control, nil
}

func (lp *LoadPolicy) GetFramework(frameworkName string) (*reporthandling.Framework, error) {
	var framework reporthandling.Framework
	var err error
	for _, filePath := range lp.filePaths {
		framework = reporthandling.Framework{}
		f, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		if err = json.Unmarshal(f, &framework); err != nil {
			return nil, err
		}
		if strings.EqualFold(frameworkName, framework.Name) {
			break
		}
	}
	if frameworkName != "" && !strings.EqualFold(frameworkName, framework.Name) {

		return nil, fmt.Errorf("framework from file not matching")
	}
	return &framework, err
}

func (lp *LoadPolicy) GetFrameworks() ([]reporthandling.Framework, error) {
	frameworks := []reporthandling.Framework{}
	var err error
	return frameworks, err
}

func (lp *LoadPolicy) ListFrameworks() ([]string, error) {
	fwNames := []string{}
	framework := &reporthandling.Framework{}
	for _, f := range lp.filePaths {
		file, err := os.ReadFile(f)
		if err == nil {
			if err := json.Unmarshal(file, framework); err == nil {
				if !contains(fwNames, framework.Name) {
					fwNames = append(fwNames, framework.Name)
				}
			}
		}
	}
	return fwNames, nil
}

func (lp *LoadPolicy) ListControls() ([]string, error) {
	// TODO - Support
	return []string{}, fmt.Errorf("loading controls list from file is not supported")
}

func (lp *LoadPolicy) GetExceptions(clusterName string) ([]armotypes.PostureExceptionPolicy, error) {
	filePath := lp.filePath()
	exception := []armotypes.PostureExceptionPolicy{}
	f, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(f, &exception)
	return exception, err
}

func (lp *LoadPolicy) GetControlsInputs(clusterName string) (map[string][]string, error) {
	filePath := lp.filePath()
	accountConfig := &armotypes.CustomerConfig{}
	f, err := os.ReadFile(filePath)
	fileName := filepath.Base(filePath)
	if err != nil {
		formattedError := fmt.Errorf("Error opening %s file, \"controls-config\" will be downloaded from ARMO management portal", fileName)
		return nil, formattedError
	}

	if err = json.Unmarshal(f, &accountConfig.Settings.PostureControlInputs); err == nil {
		return accountConfig.Settings.PostureControlInputs, nil
	}

	formattedError := fmt.Errorf("Error reading %s file, %s, \"controls-config\" will be downloaded from ARMO management portal", fileName, err.Error())

	return nil, formattedError
}

// temporary support for a list of files
func (lp *LoadPolicy) filePath() string {
	if len(lp.filePaths) > 0 {
		return lp.filePaths[0]
	}
	return ""
}

func (lp *LoadPolicy) GetAttackTracks() ([]v1alpha1.AttackTrack, error) {
	attackTracks := []v1alpha1.AttackTrack{}

	f, err := os.ReadFile(lp.filePath())

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(f, &attackTracks); err != nil {
		return nil, err
	}
	return attackTracks, nil
}
