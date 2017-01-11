package tools

import (
	"os"

	"github.com/sapk/sca/pkg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

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
	if verbose, _ := cmd.Flags().GetBool(pkg.VerboseFlag); verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
