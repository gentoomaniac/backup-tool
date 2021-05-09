package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/gentoomaniac/backup-tool/pkg/db"
	"github.com/manifoldco/promptui"
)

// PromptInstanceSelection Create a prompt that will display all instances for a user to select one
func PromptBackups(backups []*db.Backup) (*db.Backup, error) {
	if len(backups) == 1 {
		return backups[0], nil
	}

	// Sort list
	sort.Slice(backups, func(i, j int) bool {
		return strings.Compare(backups[i].Name, backups[j].Name) < 0
	})

	instanceSearchFunc := func(input string, idx int) bool {
		backup := backups[idx]

		return strings.Contains(strings.ToLower(backup.Name), strings.ToLower(input))
	}

	size := len(backups)
	if size >= 10 {
		size = 10
	}

	selector := promptui.Select{
		Label:             "Select the backup to restore",
		Items:             backups,
		Searcher:          instanceSearchFunc,
		StartInSearchMode: true,
		HideSelected:      true,
		Size:              size,
		Templates: &promptui.SelectTemplates{
			Active:   fmt.Sprintf("%s {{ .Name | cyan }}", promptui.IconSelect),
			Inactive: " {{ .Name }}",
			Details: `
{{ "Details:" | bold }}
	{{ "Name:" | bold }}	{{ .Name | cyan }}
	{{ "Description:" | bold }}	{{ .Description | cyan }}
	{{ "Created:" | bold }}	{{ .Timestamp | cyan }}
	{{ "Expiry:" | bold }}	{{ .Expiration | cyan }}
`,
			Selected: "{{ .Name }}",
		},
	}

	// Nice hack to be able to export the result to env vars
	// export $(./awstool -print-exportable=true)
	selector.Stdout = os.Stderr

	index, _, err := selector.Run()
	if err != nil {
		os.Stdout.Sync()
		return nil, err
	}

	return backups[index], nil
}
