package gitconfig

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

func ignoreGitSystemConfig() func() {
	const GIT_CONFIG_NOSYSTEM = "GIT_CONFIG_NOSYSTEM"

	prevGitConfigNoSystem, wasSet := os.LookupEnv(GIT_CONFIG_NOSYSTEM)
	os.Setenv(GIT_CONFIG_NOSYSTEM, "")

	return func (){
		os.Unsetenv(GIT_CONFIG_NOSYSTEM)
		if wasSet {
			os.Setenv(GIT_CONFIG_NOSYSTEM, prevGitConfigNoSystem)
		}
	}
}

func withTemporaryGitGlobalConfigDirectory() (string, func()) {
	tmpdir, err := ioutil.TempDir("", "go-gitconfig-test-global")
	if err != nil {
		panic(err)
	}

	const HOME = "HOME"
	prevHome, homeWasSet := os.LookupEnv(HOME)
	os.Setenv(HOME, tmpdir)

	const XDG_CONFIG_HOME = "XDG_CONFIG_HOME"
	prevXdgConfigHome, xdgWasSet := os.LookupEnv(XDG_CONFIG_HOME)
	os.Setenv(XDG_CONFIG_HOME, tmpdir)

	return tmpdir, func() {
		os.Unsetenv(XDG_CONFIG_HOME)
		if xdgWasSet {
			os.Setenv(XDG_CONFIG_HOME, prevXdgConfigHome)
		}

		os.Unsetenv(HOME)
		if homeWasSet {
			os.Setenv(HOME, prevHome)
		}
		os.RemoveAll(tmpdir)
	}
}

func withTemporaryGitLocalConfigDirectory() func() {
	prevDir, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}

	tmpdir, err := ioutil.TempDir("", "go-gitconfig-test-local")
	if err != nil {
		panic(err)
	}

	os.Chdir(tmpdir)

	gitInit := exec.Command("git", "init")
	gitInit.Stderr = ioutil.Discard
	if err = gitInit.Run(); err != nil {
		panic(err)
	}

	return func() {
		os.Chdir(prevDir)
		os.RemoveAll(tmpdir)
	}

}

func withGlobalGitConfigFile(content string) func() {
	resetIgnore := ignoreGitSystemConfig()

	tmpGlobalDir, resetGlobal := withTemporaryGitGlobalConfigDirectory()
	resetLocal := withTemporaryGitLocalConfigDirectory()

	tmpGitConfigFile := filepath.Join(tmpGlobalDir, ".gitconfig")

	ioutil.WriteFile(
		tmpGitConfigFile,
		[]byte(content),
		0777,
	)

	return func() {
		resetLocal()
		resetGlobal()
		resetIgnore()
	}
}

func includeGitConfigFile(content string) (string, func()) {
	tmpdir, err := ioutil.TempDir("", "go-gitconfig-test-include")
	if err != nil {
		panic(err)
	}

	tmpGitIncludeConfigFile := filepath.Join(tmpdir, ".gitconfig.local")
	ioutil.WriteFile(
		tmpGitIncludeConfigFile,
		[]byte(content),
		0777,
	)

	return tmpGitIncludeConfigFile, func() {
		os.RemoveAll(tmpdir)
	}
}

func withLocalGitConfigFile(key string, value string) func() {
	var err error

	resetIgnore := ignoreGitSystemConfig()

	_, resetGlobal := withTemporaryGitGlobalConfigDirectory()
	resetLocal := withTemporaryGitLocalConfigDirectory()

	gitAddConfig := exec.Command("git", "config", "--local", key, value)
	gitAddConfig.Stderr = ioutil.Discard
	if err = gitAddConfig.Run(); err != nil {
		panic(err)
	}

	return func() {
		resetLocal()
		resetGlobal()
		resetIgnore()
	}
}
