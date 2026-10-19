package cortex

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
)

// SurrealDB credential resolution (F19). There is deliberately no built-in
// default password: the previous root/root fallback meant every default
// install shared one publicly known credential.
//
// Resolution order for NewSurrealStore and the managed Sidecar:
//  1. credentials supplied by the caller (SurrealConfig / Sidecar fields)
//  2. MIMIR_SURREAL_USER / MIMIR_SURREAL_PASS environment variables
//  3. for a store connecting to a managed sidecar address, the per-run
//     password the Sidecar generated at launch (resolveSidecarCreds)
//  4. otherwise an error: the caller must configure credentials
//
// defaultSurrealUser is a username only (the conventional name of the
// SurrealDB root account). It is never used as a password.
const defaultSurrealUser = "root"

const (
	envSurrealUser = "MIMIR_SURREAL_USER"
	envSurrealPass = "MIMIR_SURREAL_PASS"
)

// sidecarCreds maps a managed sidecar's HTTP base URL to the credentials it
// was launched with, so a SurrealStore created later in the same process can
// authenticate to it. Values live only in process memory and are never
// logged or persisted.
var sidecarCreds sync.Map // string addr -> [2]string{user, pass}

func registerSidecarCreds(httpAddr, user, pass string) {
	sidecarCreds.Store(strings.TrimRight(httpAddr, "/"), [2]string{user, pass})
}

func lookupSidecarCreds(httpAddr string) (user, pass string, ok bool) {
	v, found := sidecarCreds.Load(strings.TrimRight(httpAddr, "/"))
	if !found {
		return "", "", false
	}
	c := v.([2]string)
	return c[0], c[1], true
}

// generateSurrealPassword returns a random per-run password for a managed
// sidecar. It is shared with the in-process client via registerSidecarCreds,
// so a zero-config local install works without any known credential existing.
func generateSurrealPassword() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate password: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// resolveSurrealCreds applies the resolution order above for a direct store
// connection to addr. It returns an error rather than inventing a credential.
func resolveSurrealCreds(user, pass, addr string) (string, string, error) {
	if user == "" {
		user = os.Getenv(envSurrealUser)
	}
	if pass == "" {
		pass = os.Getenv(envSurrealPass)
	}
	if pass == "" {
		if su, sp, ok := lookupSidecarCreds(addr); ok {
			if user == "" {
				user = su
			}
			pass = sp
		}
	}
	if pass == "" {
		return "", "", fmt.Errorf("surreal: no credentials for %s: set %s and %s, pass credentials in the config, or connect to a managed sidecar", addr, envSurrealUser, envSurrealPass)
	}
	if user == "" {
		user = defaultSurrealUser
	}
	return user, pass, nil
}
