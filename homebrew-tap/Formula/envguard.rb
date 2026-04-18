class Envguard < Formula
  desc "A lightweight CLI to keep your .env files honest"
  homepage "https://github.com/BLemine/envguard"
  version "0.2.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/BLemine/envguard/releases/download/v0.2.0/envguard-darwin-arm64"
      sha256 "b4568785967ad1aa178fb08b683c2be9e3b81b5db021684b7722dadad88c35dd"
    else
      url "https://github.com/BLemine/envguard/releases/download/v0.2.0/envguard-darwin-amd64"
      sha256 "cc90655d43805385e0c552ff041ea1a5dbaef854c6b97b81cb058650c33e3cd7"
    end
  end

  on_linux do
    url "https://github.com/BLemine/envguard/releases/download/v0.2.0/envguard-linux-amd64"
    sha256 "5668b1d3c55a9c039371a0a740e6c0266e48635d5420785c8295ecba492022f5"
  end

  def install
    if OS.mac? && Hardware::CPU.arm?
      bin.install "envguard-darwin-arm64" => "envguard"
    elsif OS.mac?
      bin.install "envguard-darwin-amd64" => "envguard"
    else
      bin.install "envguard-linux-amd64" => "envguard"
    end
  end

  test do
    system "#{bin}/envguard", "--help"
  end
end
