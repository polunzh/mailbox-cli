package release

import (
	"errors"
	"fmt"
	"strings"
)

type HomebrewFormulaData struct {
	Repo           string
	Version        string
	DarwinAMD64SHA string
	DarwinARM64SHA string
}

func RenderHomebrewFormula(data HomebrewFormulaData) (string, error) {
	if strings.TrimSpace(data.Repo) == "" {
		return "", errors.New("repo is required")
	}
	if strings.TrimSpace(data.Version) == "" {
		return "", errors.New("version is required")
	}
	if strings.TrimSpace(data.DarwinAMD64SHA) == "" {
		return "", errors.New("darwin amd64 sha256 is required")
	}
	if strings.TrimSpace(data.DarwinARM64SHA) == "" {
		return "", errors.New("darwin arm64 sha256 is required")
	}

	version := strings.TrimPrefix(strings.TrimSpace(data.Version), "v")
	tag := strings.TrimSpace(data.Version)
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + version
	}

	homepage := fmt.Sprintf("https://github.com/%s", data.Repo)
	armURL := fmt.Sprintf("%s/releases/download/%s/mailbox_%s_darwin-arm64.tar.gz", homepage, tag, version)
	intelURL := fmt.Sprintf("%s/releases/download/%s/mailbox_%s_darwin-amd64.tar.gz", homepage, tag, version)

	formula := fmt.Sprintf(`class Mailbox < Formula
  desc "Terminal email client for developers"
  homepage "%s"
  version "%s"
  license "MIT"

  on_macos do
    on_arm do
      url "%s"
      sha256 "%s"
    end

    on_intel do
      url "%s"
      sha256 "%s"
    end
  end

  def install
    bin.install "mailbox"
    prefix.install "README.md", "README.zh-CN.md"
  end

  test do
    system "#{bin}/mailbox", "--help"
  end
end
`, homepage, version, armURL, data.DarwinARM64SHA, intelURL, data.DarwinAMD64SHA)

	return formula, nil
}
