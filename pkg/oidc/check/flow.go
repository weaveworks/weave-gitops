package check

import (
	"context"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

type Options struct {
	ClientID                   string
	ClientSecret               string
	IssuerURL                  string
	SecretName                 string
	SecretNamespace            string
	Scopes                     []string
	ClaimUsername              string
	ClaimGroups                string
	OpenURL                    func(string) error
	InsecureSkipSignatureCheck bool
}

type Claims struct {
	Username string
	Groups   []string
}

// GetPrincipal retrieves OIDC claims by sending the user through an authorization code flow. It spins
// up a temporary web server, sets the server's address as redirect URI in the authentication request
// and subsequently exchanges the authorization code for an ID token.
// NOTE: Make sure to configure your OIDC provider so that it accepts "http://localhost:9876" as redirect URI.
func GetPrincipal(ctx context.Context, opts Options, log logger.Logger, c client.Client) (*auth.UserPrincipal, error) {
	if opts.SecretName != "" {
		if err := optsFromSecret(ctx, &opts, log, c); err != nil {
			return nil, fmt.Errorf("failed reading options from Secret: %w", err)
		}
	}

	if opts.ClaimUsername == "" {
		opts.ClaimUsername = auth.ClaimUsername
	}

	if opts.ClaimGroups == "" {
		opts.ClaimGroups = auth.ClaimGroups
	}

	if len(opts.Scopes) == 0 {
		opts.Scopes = auth.DefaultScopes
	}

	provider, err := oidc.NewProvider(ctx, opts.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("could not create provider: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		RedirectURL:  "http://localhost:9876",
		Endpoint:     provider.Endpoint(),
		Scopes:       opts.Scopes,
	}

	log.Waitingf("Opening browser. If this does not work, please open the following URL in your browser:\n")
	authCodeURL := oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	log.Println("%s\n", authCodeURL)
	var openErr error
	if opts.OpenURL != nil {
		openErr = opts.OpenURL(authCodeURL)
	} else {
		openErr = browser.OpenURL(authCodeURL)
	}
	if openErr != nil {
		log.Failuref("Failed to open browser: %s. You can still open the URL manually.", openErr)
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID:                   opts.ClientID,
		InsecureSkipSignatureCheck: opts.InsecureSkipSignatureCheck,
	})
	claims, err := retrieveIDToken(log, oauth2Config, verifier)
	if err != nil {
		return nil, fmt.Errorf("failed retrieving claims: %w", err)
	}

	cc := auth.ClaimsConfig{
		Username: opts.ClaimUsername,
		Groups:   opts.ClaimGroups,
	}
	principal, err := cc.PrincipalFromClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("failed deriving principal from claims: %w", err)
	}

	return principal, nil
}

// optsFromSecret fetches the Secret referenced in opts and sets OIDC configuration values
// from that Secret. It will not override a value if it is already set to non-empty in opts but
// will rather only fill in empty values.
func optsFromSecret(ctx context.Context, opts *Options, log logger.Logger, c client.Client) error {
	oidcSecret := corev1.Secret{}

	missingFields := make([]string, 0)

	ref := types.NamespacedName{
		Namespace: opts.SecretNamespace,
		Name:      opts.SecretName,
	}
	log.Actionf("Fetching OIDC configuration from Secret %q", ref)
	if err := c.Get(ctx, ref, &oidcSecret); err != nil {
		return fmt.Errorf("failed retrieving Secret %q from cluster: %w", ref, err)
	}

	if opts.ClientID == "" {
		opts.ClientID = string(oidcSecret.Data["clientID"])
		if opts.ClientID == "" {
			missingFields = append(missingFields, "clientID")
		}
	}
	if opts.ClientSecret == "" {
		opts.ClientSecret = string(oidcSecret.Data["clientSecret"])
		if opts.ClientSecret == "" {
			missingFields = append(missingFields, "clientSecret")
		}
	}
	if opts.IssuerURL == "" {
		opts.IssuerURL = string(oidcSecret.Data["issuerURL"])
		if opts.IssuerURL == "" {
			missingFields = append(missingFields, "issuerURL")
		}
	}
	if opts.ClaimUsername == "" {
		opts.ClaimUsername = string(oidcSecret.Data["claimUsername"])
	}
	if len(opts.Scopes) == 0 {
		cs := string(oidcSecret.Data["customScopes"])
		if cs != "" {
			opts.Scopes = strings.Split(cs, ",")
		}
	}

	// only check for existence of this field for proper feedback
	if _, ok := oidcSecret.Data["redirectURL"]; !ok {
		missingFields = append(missingFields, "redirectURL")
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("secret is missing fields: %s", strings.Join(missingFields, ","))
	}

	return nil
}
