package templates

import (
	"fmt"
	"io"
	"strings"
)

// TemplatesRetriever defines the interface that adapters
// need to implement in order to return an array of templates.
type TemplatesRetriever interface {
	Source() string
	RetrieveTemplates() ([]Template, error)
	RetrieveTemplateParameters(name string) ([]TemplateParameter, error)
}

type Template struct {
	Name        string
	Description string
}

type TemplateParameter struct {
	Name        string
	Description string
	Required    bool
	Options     []string
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

// GetTemplateParameters uses a TemplatesRetriever adapter
// to show a list of parameters for a given template.
func GetTemplateParameters(name string, r TemplatesRetriever, w io.Writer) error {
	ps, err := r.RetrieveTemplateParameters(name)
	if err != nil {
		return fmt.Errorf("unable to retrieve parameters for template %q from %q: %w", name, r.Source(), err)
	}

	if len(ps) > 0 {
		fmt.Fprintf(w, "NAME\tREQUIRED\tDESCRIPTION\tOPTIONS\n")

		for _, t := range ps {
			fmt.Fprintf(w, "%s", t.Name)
			fmt.Fprintf(w, "\t%t", t.Required)

			if t.Description != "" {
				fmt.Fprintf(w, "\t%s", t.Description)
			}

			if t.Options != nil {
				optionsStr := strings.Join(t.Options, ", ")
				fmt.Fprintf(w, "\t%s", optionsStr)
			}

			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No template parameters were found.")

	return nil
}
