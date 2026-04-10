# This file is auto-updated by GoReleaser on each release.
# Do not edit manually.
class FractalxCli < Formula
  desc "FractalX CLI — generate a Spring Boot monolith ready for decomposition"
  homepage "https://github.com/fractalx-org/fractalx-cli"
  version "0.1.0"
  license "Apache-2.0"

  on_macos do
    on_arm do
      url "https://github.com/fractalx-org/fractalx-cli/releases/download/v#{version}/fractalx-cli_#{version}_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM64"
    end
    on_intel do
      url "https://github.com/fractalx-org/fractalx-cli/releases/download/v#{version}/fractalx-cli_#{version}_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_AMD64"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/fractalx-org/fractalx-cli/releases/download/v#{version}/fractalx-cli_#{version}_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_ARM64"
    end
    on_intel do
      url "https://github.com/fractalx-org/fractalx-cli/releases/download/v#{version}/fractalx-cli_#{version}_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER_LINUX_AMD64"
    end
  end

  def install
    bin.install "fractalx"
  end

  test do
    system "#{bin}/fractalx", "--help"
  end
end
