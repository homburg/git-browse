package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

///
/// TODO: Use git.go from homburg/lab as git library instead
/// copying go files.
/// Convert server type to "remote"
///

type gitRemote struct {
	base string
	path string
}

type gitDir string

type server struct {
	scheme string
	host   string
}

func newServer(host string) server {
	return server{"http", host}
}

func (s server) getProjectUrl(path string) string {
	return s.scheme + "://" + s.host + "/" + strings.TrimPrefix(path, "/")
}

func parseRemote(remoteAddr string) (remote gitRemote) {
	// Strip user info
	if atIndex := strings.IndexByte(remoteAddr, '@'); atIndex >= 0 {
		remoteAddr = remoteAddr[atIndex+1 : len(remoteAddr)]
	}

	if schemeIndex := strings.Index(remoteAddr, "://"); schemeIndex >= 0 {
		remoteAddr = remoteAddr[schemeIndex+3 : len(remoteAddr)]
	}

	if i := strings.IndexAny(remoteAddr, ":/"); i >= 0 {
		remote = gitRemote{
			remoteAddr[0:i],
			remoteAddr[i+1 : len(remoteAddr)],
		}
	} else {
		// relative remote? Not interested anyway
	}

	remote.path = strings.TrimSuffix(remote.path, ".git")

	return remote
}

/// Get remote url or fail!
func needRemoteURL(remote string) gitRemote {
	// remote := c.String("remote")
	git := needGitDir("")
	remoteURL, err := git.getRemoteURL(remote)
	if nil != err {
		log.Fatal(err)
	}

	return parseRemote(remoteURL)
}

/// Get gitlab url or fail!
func needServer() server {
	r := needRemoteURL("origin")
	return newServer(r.base)
}

/// Get origin for given remote name
func (here gitDir) getRemoteURL(remoteName string) (string, error) {
	cmd := exec.Command("git", "--git-dir", string(here), "remote", "-v")
	output, err := cmd.CombinedOutput()

	if nil != err {
		return "", fmt.Errorf("%s\n", output)
	}

	return getRemoteURLFromRemoteVOutput(remoteName, output)
}

func needGitDir(given string) gitDir {
	// given := c.String("git-dir")
	var err error
	if given == "" {
		given, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
	}

	return gitDir(strings.TrimSuffix(given, "/") + "/.git")
}

type ErrUnknownRemote string

func (e ErrUnknownRemote) Error() string {
	return fmt.Sprintf("Could not find remote: %s\n", e)
}

func getRemoteURLFromRemoteVOutput(remoteName string, output []byte) (string, error) {
	lines := strings.Split(string(output), "\n")
	prefix := remoteName + "\t"
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			remoteLine := strings.TrimPrefix(line, prefix)
			remoteLine = strings.TrimSuffix(remoteLine, " (fetch)")
			remoteLine = strings.TrimSuffix(remoteLine, " (push)")
			return remoteLine, nil
		}
	}

	return "", ErrUnknownRemote(remoteName)
}
