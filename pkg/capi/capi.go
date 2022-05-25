package capi

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
	RetrieveTemplatesByProvider(provider string) ([]Template, error)
	RetrieveTemplateParameters(name string) ([]TemplateParameter, error)
	RetrieveTemplateProfiles(name string) ([]Profile, error)
}

// TemplateRenderer defines the interface that adapters
// need to implement in order to render a template populated
// with parameter values.
type TemplateRenderer interface {
	RenderTemplateWithParameters(name string, parameters map[string]string, creds Credentials) (string, error)
}

// CredentialsRetriever defines the interface that adapters
// need to implement in order to retrieve CAPI credentials.
type CredentialsRetriever interface {
	Source() string
	RetrieveCredentials() ([]Credentials, error)
}

type Template struct {
	Name        string
	Description string
	Provider    string
	Error       string
}

type TemplateParameter struct {
	Name        string
	Description string
	Required    bool
	Options     []string
}

type Credentials struct {
	Group     string `json:"group"`
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type ProfileValues struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Values  string `json:"values"`
}

type Profile struct {
	Name              string
	Home              string
	Sources           []string
	Description       string
	Keywords          []string
	Maintainers       []Maintainer
	Icon              string
	Annotations       map[string]string
	KubeVersion       string
	HelmRepository    HelmRepository
	AvailableVersions []string
}

type HelmRepository struct {
	Name      string
	Namespace string
}

type Maintainer struct {
	Name  string
	Email string
	Url   string
}

// GetTemplates uses a TemplatesRetriever adapter to show
// a list of templates to the console.
func GetTemplates(r TemplatesRetriever, w io.Writer) error {
	ts, err := r.RetrieveTemplates()
	if err != nil {
		return fmt.Errorf("unable to retrieve templates from %q: %w", r.Source(), err)
	}

	if len(ts) > 0 {
		fmt.Fprintf(w, "NAME\tPROVIDER\tDESCRIPTION\tERROR\n")

		for _, t := range ts {
			fmt.Fprintf(w, "%s", t.Name)
			fmt.Fprintf(w, "\t%s", t.Provider)
			fmt.Fprintf(w, "\t%s", t.Description)
			fmt.Fprintf(w, "\t%s", t.Error)
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No templates were found.\n")

	return nil
}

// GetTemplatesByProvider uses a TemplatesRetriever adapter to show
// a list of templates for a given provider to the console.
func GetTemplatesByProvider(provider string, r TemplatesRetriever, w io.Writer) error {
	ts, err := r.RetrieveTemplatesByProvider(provider)
	if err != nil {
		return fmt.Errorf("unable to retrieve templates from %q: %w", r.Source(), err)
	}

	if len(ts) > 0 {
		fmt.Fprintf(w, "NAME\tPROVIDER\tDESCRIPTION\tERROR\n")

		for _, t := range ts {
			fmt.Fprintf(w, "%s", t.Name)
			fmt.Fprintf(w, "\t%s", t.Provider)
			fmt.Fprintf(w, "\t%s", t.Description)
			fmt.Fprintf(w, "\t%s", t.Error)
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No templates were found for provider %q.\n", provider)

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

	fmt.Fprintf(w, "No template parameters were found.\n")

	return nil
}

// RenderTemplate uses a TemplateRenderer adapter to show
// a template populated with parameter values.
func RenderTemplateWithParameters(name string, parameters map[string]string, creds Credentials, r TemplateRenderer, w io.Writer) error {
	t, err := r.RenderTemplateWithParameters(name, parameters, creds)
	if err != nil {
		return fmt.Errorf("unable to render template %q: %w", name, err)
	}

	if t != "" {
		fmt.Fprint(w, t)
		return nil
	}

	fmt.Fprintf(w, "No template was found.\n")

	return nil
}

// GetCredentials uses a CredentialsRetriever adapter to show
// a list of CAPI credentials.
func GetCredentials(r CredentialsRetriever, w io.Writer) error {
	cs, err := r.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("unable to retrieve credentials from %q: %w", r.Source(), err)
	}

	if len(cs) > 0 {
		fmt.Fprintf(w, "NAME\tINFRASTRUCTURE PROVIDER\n")

		for _, c := range cs {
			fmt.Fprintf(w, "%s", c.Name)
			// Extract the infra provider name from ClusterKind
			provider := c.Kind[:strings.Index(c.Kind, "Cluster")]
			fmt.Fprintf(w, "\t%s", provider)
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No credentials were found.\n")

	return nil
}

// GetTemplateProfiles uses a TemplatesRetriever adapter
// to show a list of profiles for a given template.
func GetTemplateProfiles(name string, r TemplatesRetriever, w io.Writer) error {
	ps, err := r.RetrieveTemplateProfiles(name)
	if err != nil {
		return fmt.Errorf("unable to retrieve profiles for template %q from %q: %w", name, r.Source(), err)
	}

	if len(ps) > 0 {
		fmt.Fprintf(w, "NAME\tLATEST_VERSIONS\n")

		for _, p := range ps {
			if len(p.AvailableVersions) > 5 {
				p.AvailableVersions = p.AvailableVersions[len(p.AvailableVersions)-5:]
			}

			latestVersions := strings.Join(p.AvailableVersions, ", ")

			fmt.Fprintf(w, "%s", p.Name)
			fmt.Fprintf(w, "\t%s", latestVersions)
			fmt.Fprintln(w, "")
		}

		return nil
	}

	fmt.Fprintf(w, "No template profiles were found.\n")

	return nil
}
