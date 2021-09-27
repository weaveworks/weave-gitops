package templates

import (
	"fmt"
	"io"
)

// TemplatesRetriever defines the interface that adapters
// need to implement in order to return an array of templates.
type TemplatesRetriever interface {
	Source() string
	RetrieveTemplates() ([]Template, error)
}

type Template struct {
	Name        string
	Description string
}

// GetTemplates uses a TemplatesRetriever adapter to show
// a list of templates to the console.
func GetTemplates(r TemplatesRetriever, w io.Writer) error {
	ts, err := r.RetrieveTemplates()
	if err != nil {
		return fmt.Errorf("unable to retrieve templates from %q: %w", r.Source(), err)
	}

	if len(ts) > 0 {
		fmt.Fprintf(w, "NAME\tDESCRIPTION\n")
		for _, t := range ts {
			fmt.Fprintf(w, "%s", t.Name)
			if t.Description != "" {
				fmt.Fprintf(w, "\t%s", t.Description)
			}
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No templates found.\n")

	return nil
}

