package cmd

import (
	"bytes"
	"os"

	"github.com/spf13/cobra"
)

var customCompletion = `__kd_parse_list()
{
    local kd_output
    if kd_output=$(kd list "$1" 2>/dev/null); then
        COMPREPLY=( $( compgen -W "${kd_output[*]}" -- "$cur" ) )
    fi
}

__custom_func() {
    case ${last_command} in
        kd_build)
            if [[ ${#nouns[@]} -eq 0 ]]; then
                __kd_parse_list "apps"
            fi
            return
            ;;
        kd_deploy)
            if [[ ${#nouns[@]} -eq 0 ]]; then
                __kd_parse_list "apps"
            elif [[ ${#nouns[@]} -eq 1 ]]; then
                __kd_parse_list "targets"
            fi
            return
            ;;
        kd_kubectl)
            if [[ ${#nouns[@]} -eq 0 ]]; then
                __kd_parse_list "targets"
            fi
            return
            ;;
        *)
            ;;
    esac
}
`

var cmdCompletion = &cobra.Command{
	Use:   "completion (bash | zsh)",
	Short: "Output shell completion code for the specified shell",
	DisableFlagsInUseLine: true,

	ValidArgs: []string{"bash", "zsh"},
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}

		if err := cobra.OnlyValidArgs(cmd, args); err != nil {
			return err
		}

		return nil
	},

	Long: `Output shell completion code for the specified shell (bash or zsh). The shell
code must be evaluated to provide interactive completion of kd commands.
This can be done by sourcing it from .bash_profile or .bashrc.`,

	Example: "  brew install bash-completion\n  echo 'source $(brew --prefix)/etc/bash_completion' >> ~/.bashrc\n  kd completion bash > $(brew --prefix)/etc/bash_completion.d/kd",

	Run: func(cmd *cobra.Command, args []string) {
		var buf bytes.Buffer

		switch args[0] {
		case "bash":
			if err := cmd.Parent().GenBashCompletion(&buf); err != nil {
				log.Fatal(err)
			}

			os.Stdout.Write(bytes.Replace(buf.Bytes(), []byte("__custom_func"), []byte("__kd_custom_func"), -1))

		case "zsh":
			if err := cmd.Parent().GenZshCompletion(&buf); err != nil {
				log.Fatal(err)
			}

			os.Stdout.Write(buf.Bytes())
		}
	},
}

func init() {
	cmdRoot.AddCommand(cmdCompletion)
}
