package pkg

import (
	"os"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

//Min tools minimum of int
func Min(A, B int) int {
	min := A
	if A > B {
		min = B
	}
	return min
}

//SortedKeys tools sort map[string]
func SortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

//TypeOrEnv parse cmd to file with env vars
func TypeOrEnv(cmd *cobra.Command, flag, envname string) string {
	val, _ := cmd.Flags().GetString(flag)
	if val == "" {
		val = os.Getenv(envname)
	}
	return val
}

//SetupLogger parse cmd for log level
func SetupLogger(cmd *cobra.Command, args []string) {
	if verbose, _ := cmd.Flags().GetBool(VerboseFlag); verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

/*
//GetHash return hash of a file
func GetHash(filePath string) (result string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
*/
/*
//ByContainerID sort class
type ByContainerID []docker.APIContainers

func (a ByContainerID) Len() int           { return len(a) }
func (a ByContainerID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByContainerID) Less(i, j int) bool { return a[i].ID < a[j].ID }

//ByImageID sort class
type ByImageID []docker.APIImages

func (a ByImageID) Len() int           { return len(a) }
func (a ByImageID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByImageID) Less(i, j int) bool { return a[i].ID < a[j].ID }

//ByVolumeID sort class
type ByVolumeID []docker.Volume

func (a ByVolumeID) Len() int           { return len(a) }
func (a ByVolumeID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVolumeID) Less(i, j int) bool { return a[i].Name < a[j].Name }

//ByNetworkID sort class
type ByNetworkID []docker.Network

func (a ByNetworkID) Len() int           { return len(a) }
func (a ByNetworkID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByNetworkID) Less(i, j int) bool { return a[i].ID < a[j].ID }
*/
